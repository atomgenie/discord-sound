package server

import (
	"bytes"
	"context"
	"discord-sound/player/requests"
	"discord-sound/utils/redis"
	"discord-sound/utils/uuid"
	"encoding/binary"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mediocregopher/radix/v4"
	"gopkg.in/hraban/opus.v2"
)

type guildType struct {
	id      string
	playing bool
}

var guilds map[string]*guildType = make(map[string]*guildType)
var guildMux sync.Mutex

// HandleMessage Handler for message
func HandleMessage(s *discordgo.Session, m *discordgo.MessageCreate) {

	if m.Author.ID == s.State.User.ID {
		return
	}

	channel, err := s.State.Channel(m.ChannelID)

	if err != nil {
		return
	}

	guild, err := s.State.Guild(channel.GuildID)

	if err != nil {
		return
	}

	guildMux.Lock()

	actualGuild := guilds[guild.ID]

	if actualGuild == nil {
		guildInstance := new(guildType)
		guildInstance.id = guild.ID
		guildInstance.playing = false
		guilds[guild.ID] = guildInstance
		actualGuild = guildInstance
	}

	guildMux.Unlock()

	if strings.HasPrefix(m.Content, "!!") {

		fmt.Println(actualGuild.playing)

		if actualGuild.playing {
			return
		}

		actualGuild.playing = true

		defer func() { actualGuild.playing = false }()

		instance := new(requests.Instance)
		instance.DoneChan = make(chan string)

		var soundKey string

		requestID := uuid.Gen()

		querySound := m.Content[3:]

		err = requests.RequestSong(querySound, requestID, instance)

		if err != nil {
			return
		}

		select {
		case key := <-instance.DoneChan:
			soundKey = key
		case <-time.After(30 * time.Second):
			fmt.Println("Timeout")
			return
		}

		for _, voice := range guild.VoiceStates {
			if voice.UserID == m.Message.Author.ID {
				err = playSound(s, guild.ID, voice.ChannelID, soundKey)

				if err != nil {
					return
				}

				break
			}
		}

	}
}

const (
	channelsConst  int = 2
	frameRateConst int = 48000
	frameSizeConst int = 960
	maxBytesConst  int = (frameSizeConst * 2) * 2
)

func convertSong(sound []byte) (*[][]byte, error) {

	buffer := make([][]int16, 0)
	soundStream := bytes.NewReader(sound)

	for {
		pcmBuf := make([]int16, frameSizeConst*channelsConst)
		err := binary.Read(soundStream, binary.LittleEndian, &pcmBuf)

		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}

		if err != nil {
			return nil, err
		}

		buffer = append(buffer, pcmBuf)
	}

	encoder, err := opus.NewEncoder(frameRateConst, channelsConst, opus.AppVoIP)

	if err != nil {
		return nil, err
	}

	opusFinalBuf := make([][]byte, 0)

	for _, pcm := range buffer {

		opusBuf := make([]byte, maxBytesConst)
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

func playSound(s *discordgo.Session, guildID string, channelID string, soundID string) error {

	data, err := loadSound(soundID)

	if err != nil {
		fmt.Println("Sound", err)
		return err
	}

	voiceChannel, err := s.ChannelVoiceJoin(guildID, channelID, false, true)

	if err != nil {
		return err
	}

	voiceChannel.Speaking(true)

	for _, buff := range *data {
		voiceChannel.OpusSend <- buff
	}

	fmt.Println("Done")

	voiceChannel.Speaking(false)
	voiceChannel.Disconnect()

	return nil
}
