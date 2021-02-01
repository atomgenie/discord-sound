package commands

import (
	"bytes"
	"context"
	"discord-sound/player/requests"
	"discord-sound/player/server/guilds"
	"discord-sound/utils/opusconfig"
	"discord-sound/utils/redis"
	"discord-sound/utils/uuid"
	"encoding/binary"
	"fmt"
	"io"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mediocregopher/radix/v4"
	"gopkg.in/hraban/opus.v2"
)

// HandlePlay Handle play command
func HandlePlay(s *discordgo.Session, m *discordgo.MessageCreate, argument string, guild *guilds.Type) {

	guild.Mux.Lock()
	playing := guild.Playing
	guild.Mux.Unlock()

	if !playing {
		g, err := s.State.Guild(guild.ID)

		if err != nil {
			return
		}

		var voiceID string = ""

		for _, voice := range g.VoiceStates {
			if voice.UserID == m.Message.Author.ID {
				voiceID = voice.ChannelID
				break
			}
		}

		if voiceID == "" {
			return
		}

		guild.Mux.Lock()
		guild.SoundChannelID = voiceID
		guild.Mux.Unlock()
	}

	addToQueue(argument, guild)

	if !playing {
		startQueue(s, guild)
	}
}

func startQueue(session *discordgo.Session, guild *guilds.Type) {
	guild.Mux.Lock()

	if guild.Playing {
		guild.Mux.Unlock()
		return
	}

	guild.Playing = true
	guild.Mux.Unlock()

	defer func() {
		guild.Mux.Lock()
		guild.Playing = false
		guild.Mux.Unlock()
	}()

	voiceChannel, err := session.ChannelVoiceJoin(guild.ID, guild.SoundChannelID, false, true)

	if err != nil {
		return
	}

	defer voiceChannel.Disconnect()

	for {
		guild.Mux.Lock()

		if len(guild.Queue) == 0 {
			guild.Mux.Unlock()
			break
		}

		querySound := guild.Queue[0]
		guild.Queue = guild.Queue[1:]
		guild.Mux.Unlock()

		instance := new(requests.Instance)
		instance.DoneChan = make(chan string)

		var soundKey string

		requestID := uuid.Gen()

		err := requests.RequestSong(querySound, requestID, instance)

		if err != nil {
			return
		}

		select {
		case key := <-instance.DoneChan:
			soundKey = key
		case <-time.After(30 * time.Second):
			fmt.Println("Timeout")
			requests.CancelRequest(requestID)
			continue
		}

		playSound(voiceChannel, guild.ID, guild.SoundChannelID, soundKey)

	}

	return
}

func addToQueue(sound string, guild *guilds.Type) {
	guild.Mux.Lock()
	guild.Queue = append(guild.Queue, sound)
	guild.Mux.Unlock()
}

func convertSong(sound []byte) (*[][]byte, error) {

	buffer := make([][]int16, 0)
	soundStream := bytes.NewReader(sound)

	for {
		pcmBuf := make([]int16, opusconfig.FrameSizeConst*opusconfig.ChannelsConst)
		err := binary.Read(soundStream, binary.LittleEndian, &pcmBuf)

		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}

		if err != nil {
			return nil, err
		}

		buffer = append(buffer, pcmBuf)
	}

	encoder, err := opus.NewEncoder(opusconfig.FrameRateConst, opusconfig.ChannelsConst, opus.AppAudio)

	if err != nil {
		return nil, err
	}

	opusFinalBuf := make([][]byte, 0)

	for _, pcm := range buffer {

		opusBuf := make([]byte, opusconfig.MaxBytesConst)
		n, err := encoder.Encode(pcm, opusBuf)

		if err != nil {
			return nil, err
		}

		opusFinalBuf = append(opusFinalBuf, opusBuf[:n])
	}

	return &opusFinalBuf, nil

}

func loadSound(soundID string) (*[][]byte, error) {

	var rawData []byte = nil

	err := redis.Client.Client.Do(context.Background(), radix.Cmd(&rawData, "GET", soundID))

	if err != nil {
		return nil, err
	}

	if rawData == nil {
		return nil, fmt.Errorf("Redis error")
	}

	return convertSong(rawData)
}

func playSound(voiceChannel *discordgo.VoiceConnection, guildID string, channelID string, soundID string) error {

	data, err := loadSound(soundID)

	if err != nil {
		fmt.Println("Sound", err)
		return err
	}

	voiceChannel.Speaking(true)

	for _, buff := range *data {
		voiceChannel.OpusSend <- buff
	}

	fmt.Println("Done")

	voiceChannel.Speaking(false)

	return nil
}
