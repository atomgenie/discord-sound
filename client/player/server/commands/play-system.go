package commands

import (
	"bytes"
	"context"
	"discord-sound/player/requests"
	"discord-sound/player/server/guilds"
	"discord-sound/utils/discord"
	"discord-sound/utils/opusconfig"
	"discord-sound/utils/redis"
	"discord-sound/utils/uuid"
	"encoding/binary"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mediocregopher/radix/v4"
	"gopkg.in/hraban/opus.v2"
)

func startQueue(session *discordgo.Session, guild *guilds.Type) {

	if guild.GetPlaying() {
		return
	}

	guild.SetPlaying(true)
	defer guild.SetPlaying(false)

	voiceChannel, err := session.ChannelVoiceJoin(guild.ID, guild.SoundChannelID, false, true)

	if err != nil {
		return
	}

	guild.SetVoiceChannel(voiceChannel)

	preload := new(preload)
	preloadEnd := make(chan int)

	go handlePreload(guild, preload, preloadEnd)

	for {
		querySound, isEmpty := guild.QueuePopFront()

		if isEmpty {
			break
		}

		playSound(session, guild.ID, guild.SoundChannelID, querySound, guild, preload)
	}

	guild.GetVoiceChannel().Disconnect()
	preloadEnd <- 1
	fmt.Println("End queue")

	return
}

func getIDFromQuery(query string) (string, string, error) {
	instance := new(requests.Instance)
	instance.DoneChan = make(chan requests.DoneStruct)

	var soundKey string
	var titleKey string

	requestID := uuid.Gen()

	err := requests.RequestSong(query, requestID, instance)

	if err != nil {
		return "", "", err
	}

	select {
	case key := <-instance.DoneChan:
		soundKey = key.YoutubeID
		titleKey = key.YoutubeTitle
	case <-time.After(60 * time.Second):
		fmt.Println("Request Timeout", requestID)
		requests.CancelRequest(requestID)
		return "", "", fmt.Errorf("Timeout")
	}

	return soundKey, titleKey, nil
}

func handlePreload(guild *guilds.Type, preload *preload, endChan chan int) {
	for {
		select {
		case <-endChan:
			return
		case <-time.After(10 * time.Second):
			break
		}

		queueLen := guild.QueueLen()

		if queueLen != 0 {
			query := guild.GetQueue()[0]

			preload.mux.Lock()
			if query.UUID == preload.queryID {
				preload.mux.Unlock()
			} else {
				preload.mux.Unlock()
				soundID, soundName, err := getIDFromQuery(query.Query)

				if err == nil {
					data, err := loadSound(soundID)

					if err == nil {
						preload.mux.Lock()
						preload.queryID = query.UUID
						preload.title = soundName
						preload.payload = data
						preload.mux.Unlock()
					}
				}
			}
		}
	}
}

func updateQueueTitle(query string, guild *guilds.Type) {
	title, err := requests.RequestTitle(query)

	if err != nil {
		return
	}

	guild.SetQueueTitle(query, title.Title, title.ID)
}

func addToQueue(sound string, guild *guilds.Type) {
	guild.QueueAppend(guilds.QueueType{Query: sound, UUID: uuid.Gen()})
	go updateQueueTitle(sound, guild)
}

func convertSongAsync(sound []byte, out chan []byte) {

	defer func(outchan chan []byte) {
		outchan <- nil
	}(out)

	soundStream := bytes.NewReader(sound)
	encoder, err := opus.NewEncoder(opusconfig.FrameRateConst, opusconfig.ChannelsConst, opus.AppAudio)

	if err != nil {
		return
	}

	for {
		pcmBuf := make([]int16, opusconfig.FrameSizeConst*opusconfig.ChannelsConst)
		err := binary.Read(soundStream, binary.LittleEndian, &pcmBuf)

		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}

		if err != nil {
			return
		}

		opusBuf := make([]byte, opusconfig.MaxBytesConst)
		n, err := encoder.Encode(pcmBuf, opusBuf)

		if err != nil {
			return
		}

		out <- opusBuf[:n]
	}
}

func convertSong(sound []byte) *[][]byte {
	opusFinalBuf := make([][]byte, 0)
	outChan := make(chan []byte, 60*10)

	go convertSongAsync(sound, outChan)

	for {
		data := <-outChan

		if data == nil {
			break
		}

		opusFinalBuf = append(opusFinalBuf, data)
	}

	return &opusFinalBuf
}

type preload struct {
	queryID string
	title   string
	payload *[][]byte
	mux     sync.Mutex
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

	return convertSong(rawData), nil
}

func loadSoundAsync(soundID string, out chan []byte) {
	var rawData []byte = nil
	err := redis.Client.Client.Do(context.Background(), radix.Cmd(&rawData, "GET", soundID))

	if err != nil || rawData == nil {
		out <- nil
		return
	}

	convertSongAsync(rawData, out)
}

func sendSound(s *discordgo.Session, guild *guilds.Type, data *[][]byte) {

	voiceChannel := guild.GetVoiceChannel()

	for i, buff := range *data {

		if i%60 == 0 {
			select {
			case <-guild.Skip:
				return
			case <-guild.PauseChan:
				continueMusic := handlePauseMusic(s, guild)
				voiceChannel = guild.GetVoiceChannel()

				if continueMusic {
					break
				} else {
					return
				}
			default:
				break

			}
		}

		select {
		case voiceChannel.OpusSend <- buff:
			break
		case <-time.After(5 * time.Second):
			continueMusic := handlePauseMusic(s, guild)
			voiceChannel = guild.GetVoiceChannel()

			if !continueMusic {
				return
			}
		}
	}
}

func sendSoundAsync(s *discordgo.Session, guild *guilds.Type, soundID string) {
	i := 0

	outChan := make(chan []byte, 60*10)

	voiceChannel := guild.GetVoiceChannel()

	go loadSoundAsync(soundID, outChan)

	for {
		i++

		if i%60 == 0 {
			select {
			case <-guild.Skip:
				return
			case <-guild.PauseChan:
				continueMusic := handlePauseMusic(s, guild)
				voiceChannel = guild.GetVoiceChannel()

				if continueMusic {
					break
				} else {
					return
				}
			default:
				break
			}
		}

		buff := <-outChan
		if buff == nil {
			break
		}

		select {
		case voiceChannel.OpusSend <- buff:
			break
		case <-time.After(5 * time.Second):
			continueMusic := handlePauseMusic(s, guild)
			voiceChannel = guild.GetVoiceChannel()

			if !continueMusic {
				return
			}
		}
	}
}

func playSound(s *discordgo.Session, guildID string, channelID string, query guilds.QueueType, guild *guilds.Type, preload *preload) error {

	var data *[][]byte = nil
	var soundName string
	var soundID string

	preload.mux.Lock()
	if preload.queryID == query.UUID {
		data = preload.payload
		soundName = preload.title
	}
	preload.mux.Unlock()

	if data == nil {
		_soundID, _soundName, err := getIDFromQuery(query.Query)

		if err != nil {
			fmt.Println("Sound", err)
			return err
		}

		soundName = _soundName
		soundID = _soundID
	}

	guild.SetNowPlaying(soundName)
	defer guild.SetNowPlaying("")

	guild.GetVoiceChannel().Speaking(true)
	defer guild.GetVoiceChannel().Speaking(false)

	if data != nil {
		sendSound(s, guild, data)
	} else {
		sendSoundAsync(s, guild, soundID)
	}

	fmt.Println("Done")

	return nil
}

func handlePauseMusic(s *discordgo.Session, guild *guilds.Type) bool {

	guild.SetPause(true)
	defer guild.SetPause(false)

	voiceChannel := guild.GetVoiceChannel()

	select {
	case payload := <-guild.ResumeChan:
		select {
		case voiceChannel.OpusSend <- make([]byte, 0):
			break
		case <-time.After(1 * time.Second):
			newVoiceChannelID, err := discord.GetVoiceChannel(s, payload.Message, guild.ID)

			if err != nil {
				break
			}

			voiceChannel.Disconnect()
			newVoiceChannel, err := s.ChannelVoiceJoin(guild.ID, newVoiceChannelID, false, true)

			if err != nil {
				break
			}

			guild.SoundChannelID = newVoiceChannelID
			guild.SetVoiceChannel(newVoiceChannel)
		}

		return true
	case <-time.After(5 * time.Minute):
		fmt.Println("Timeout resume", guild.ID)
		return false
	}
}
