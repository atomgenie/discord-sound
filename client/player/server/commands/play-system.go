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

// Entrypoint to start the queue system
func startQueue(session *discordgo.Session, guild *guilds.Type) {

	// Check if already playing, since we only want one thread to play music
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

	// Preload system initialisation
	preload := new(preload)
	preloadEnd := make(chan int)

	go handlePreload(guild, preload, preloadEnd)

	// Loop over musics in queue
	for {
		querySound, isEmpty := guild.QueuePopFront()

		if isEmpty {
			break
		}

		playSound(session, guild.ID, guild.SoundChannelID, querySound, guild, preload)

		// If looping over queue is activated, push the played sound in queue at the end
		if guild.GetLoop() == guilds.LoopQueue {
			guild.QueueAppend(querySound)
		}
	}

	guild.GetVoiceChannel().Disconnect()

	// Stop preload system
	preloadEnd <- 1
	fmt.Println("End queue")

	return
}

// Get id from a query over the youtubedl microservice
// Process it async
func getIDFromQuery(query string) (string, string, error) {
	// This will be used when the youtubedl service will done the download
	instance := new(requests.Instance)
	instance.DoneChan = make(chan requests.DoneStruct)

	var soundKey string
	var titleKey string

	// Generate an unique ID to track the request, and check who is the requester when system will receive the response
	requestID := uuid.Gen()

	err := requests.RequestSong(query, requestID, instance)

	if err != nil {
		return "", "", err
	}

	select {
	// We got a response from our service
	case key := <-instance.DoneChan:
		soundKey = key.YoutubeID
		titleKey = key.YoutubeTitle

	// The service has tiemout
	case <-time.After(60 * time.Second):
		fmt.Println("Request Timeout", requestID)
		requests.CancelRequest(requestID)
		return "", "", fmt.Errorf("Timeout")
	}

	return soundKey, titleKey, nil
}

// Preload system
// Will try to get music data for the next sound in queue
func handlePreload(guild *guilds.Type, preload *preload, endChan chan int) {
	for {
		select {
		// Check if we should stop the preload system
		case <-endChan:
			return
		// Preload the next song every 10 seconds (it's a sleep)
		case <-time.After(10 * time.Second):
			break
		}

		queueLen := guild.QueueLen()

		if queueLen != 0 {
			query := guild.GetQueue()[0]

			preload.mux.Lock()
			// Do not preload if music is already preloaded
			if query.UUID == preload.queryID {
				preload.mux.Unlock()
			} else {
				preload.mux.Unlock()

				// We request here the data of the music
				soundID, soundName, err := getIDFromQuery(query.Query)

				if err == nil {
					data, err := preloadSound(soundID)

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

// Set the Youtube Title of a query in the queue
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

// Async convert the song to Discord compatible format (opus)
func convertSongAsync(sound []byte, out chan []byte) {

	// Send nil to channel at the end of music
	defer func(outchan chan []byte) {
		outchan <- nil
	}(out)

	soundStream := bytes.NewReader(sound)
	encoder, err := opus.NewEncoder(opusconfig.FrameRateConst, opusconfig.ChannelsConst, opus.AppAudio)

	if err != nil {
		return
	}

	for {

		// Convert to PCM
		pcmBuf := make([]int16, opusconfig.FrameSizeConst*opusconfig.ChannelsConst)
		err := binary.Read(soundStream, binary.LittleEndian, &pcmBuf)

		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}

		if err != nil {
			return
		}

		// Convert to opus
		opusBuf := make([]byte, opusconfig.MaxBytesConst)
		n, err := encoder.Encode(pcmBuf, opusBuf)

		if err != nil {
			return
		}

		out <- opusBuf[:n]
	}
}

type preload struct {
	queryID string
	title   string
	payload *[]byte
	mux     sync.Mutex
}

// Only get data from youtubedl microservice (via Redis) but not convert it to opus
func preloadSound(soundID string) (*[]byte, error) {
	var rawData []byte = make([]byte, 0)
	// The data is already set by the youtubedl microservice
	// This need to be triggered by calling the microservice before
	err := redis.Client.Client.Do(context.Background(), radix.Cmd(&rawData, "GET", soundID))

	if err != nil {
		return nil, err
	}

	if rawData == nil {
		return nil, fmt.Errorf("Redis error")
	}

	return &rawData, nil
}

func loadSoundAsync(soundID string, queryID string, preload *preload, out chan []byte) {
	var rawData []byte = nil

	preload.mux.Lock()
	preloadID := preload.queryID
	preload.mux.Unlock()

	// If music is already preloaded, retrive data from preload system, else get it from redis
	if preloadID == queryID {
		rawData = *preload.payload
	} else {
		err := redis.Client.Client.Do(context.Background(), radix.Cmd(&rawData, "GET", soundID))
		if err != nil || rawData == nil {
			out <- nil
			return
		}
	}

	convertSongAsync(rawData, out)
}

// Send sound to discord channel
func sendSoundAsync(s *discordgo.Session, guild *guilds.Type, preload *preload, queryID string, soundID string) {

	// If loop over music is activated
	for guild.GetLoop() == guilds.LoopSound {

		i := 0
		outChan := make(chan []byte, 60*10)

		// Get voice channel (can change over a run)
		voiceChannel := guild.GetVoiceChannel()

		// Load sound from Redis, convert it and put result in outChan
		go loadSoundAsync(soundID, queryID, preload, outChan)

		for {
			i++

			// Check if some commands are triggered by Discord message (skip, pause...)
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

			// If it is the end of the music
			if buff == nil {
				break
			}

			// If we can't send data, activate pause
			// Occur if we move the bot in another channel
			// We can resume the music
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
}

func playSound(s *discordgo.Session, guildID string, channelID string, query guilds.QueueType, guild *guilds.Type, preload *preload) error {

	var soundName string
	var soundID string
	isPreload := false

	preload.mux.Lock()
	if preload.queryID == query.UUID {
		soundName = preload.title
		isPreload = true
	}
	preload.mux.Unlock()

	if !isPreload {
		_soundID, _soundName, err := getIDFromQuery(query.Query)

		if err != nil {
			fmt.Println("Sound", err)
			return err
		}

		soundName = _soundName
		soundID = _soundID
	}

	guild.SetNowPlaying(soundName, soundID)
	defer guild.SetNowPlaying("", "")

	guild.GetVoiceChannel().Speaking(true)
	defer guild.GetVoiceChannel().Speaking(false)

	sendSoundAsync(s, guild, preload, query.UUID, soundID)

	fmt.Println("Done")

	return nil
}

// Activate pause
func handlePauseMusic(s *discordgo.Session, guild *guilds.Type) bool {
	// Activate pause
	guild.SetPause(true)
	defer guild.SetPause(false)

	voiceChannel := guild.GetVoiceChannel()

	select {
	// Resume is aasked by users
	case payload := <-guild.ResumeChan:
		select {
		// Verify if we can send data over the voice channel
		case voiceChannel.OpusSend <- make([]byte, 0):
			break
		// Force reconnection if not
		case <-time.After(1 * time.Second):

			// Get the voice channel where the user who has send the message is connected to
			newVoiceChannelID, err := discord.GetVoiceChannel(s, payload.Message, guild.ID)

			if err != nil {
				break
			}

			// Disconnect and reconnect bot
			voiceChannel.Disconnect()
			newVoiceChannel, err := s.ChannelVoiceJoin(guild.ID, newVoiceChannelID, false, true)

			if err != nil {
				break
			}

			// Update current voice channel
			guild.SoundChannelID = newVoiceChannelID
			guild.SetVoiceChannel(newVoiceChannel)
		}

		return true

	// Leave bot if pause is too long
	// TODO: Handle this, because it will stop the music, not the whole queue
	case <-time.After(5 * time.Minute):
		fmt.Println("Timeout resume", guild.ID)
		return false
	}
}
