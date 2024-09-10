package main

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

// Variables globales para permisos y datos
var permissions []interface{}
var permissionsData map[string]interface{}

// Cargar permisos desde info.json
func loadPermissions() (map[string]interface{}, error) {
	file, err := os.Open("info.json")
	if err != nil {
		return nil, fmt.Errorf("error abriendo info.json: %v", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var data map[string]interface{}
	err = decoder.Decode(&data)
	if err != nil {
		return nil, fmt.Errorf("error decodificando info.json: %v", err)
	}

	return data, nil
}

// Comprobar si el usuario tiene permisos
func hasPermission(userID string, permissions []interface{}) bool {
	for _, id := range permissions {
		if id == userID {
			return true
		}
	}
	return false
}

func main() {
	// Cargar permisos desde info.json
	var err error
	permissionsData, err = loadPermissions()
	if err != nil {
		log.Fatalf("Error cargando permisos: %v", err)
	}

	// Obtener la lista de usuarios con permiso
	permissions = permissionsData["permissions"].([]interface{})

	// Cargar variables de entorno desde el archivo .env
	err = godotenv.Load()
	if err != nil {
		log.Fatalf("Error cargando el archivo .env: %v", err)
	}

	// Obtener el token de Discord desde las variables de entorno
	DISCORD_TOKEN := os.Getenv("DISCORD_TOKEN")
	if DISCORD_TOKEN == "" {
		log.Fatalf("DISCORD_TOKEN no está definido en el archivo .env")
	}

	// Crear una nueva sesión de Discord usando el token del bot
	dg, err := discordgo.New("Bot " + DISCORD_TOKEN)
	if err != nil {
		fmt.Println("Error creando sesión de Discord,", err)
		return
	}

	// Activar intents para recibir eventos de mensajes
	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages

	// Registrar una función de callback para cuando el bot reciba un mensaje
	dg.AddHandler(messageCreate)

	// Abrir una conexión WebSocket al servidor de Discord
	err = dg.Open()
	if err != nil {
		fmt.Println("Error abriendo conexión,", err)
		return
	}

	fmt.Println("Bot está corriendo. Presiona CTRL+C para salir.")

	// Esperar hasta que se interrumpa el programa (CTRL+C)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-stop

	// Cerrar la sesión cuando termine el programa
	dg.Close()
}

// Función que se llama cada vez que el bot recibe un mensaje
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignorar mensajes enviados por el mismo bot
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Verificar permisos del usuario
	if len(m.Content) > 0 && m.Content[0] == '!' && !hasPermission(m.Author.ID, permissions) {
		s.ChannelMessageDelete(m.ChannelID, m.ID)
		s.ChannelMessageSend(m.ChannelID, "No tienes permiso para usar el bot")
		return
	}

	// Respuesta a comando !ping
	if m.Content == "!ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
		err := s.ChannelMessageDelete(m.ChannelID, m.ID)
		if err != nil {
			fmt.Println("Error al eliminar el mensaje:", err)
		}
		return
	}

	// Comando para limpiar mensajes del canal
	if m.Content == "!clean" {
		messages, err := s.ChannelMessages(m.ChannelID, 100, "", "", "")
		if err != nil {
			log.Println("Error obteniendo mensajes:", err)
			return
		}

		var messageIDs []string
		for _, msg := range messages {
			messageIDs = append(messageIDs, msg.ID)
		}

		if len(messageIDs) > 0 {
			err = s.ChannelMessagesBulkDelete(m.ChannelID, messageIDs)
			if err != nil {
				log.Println("Error eliminando mensajes:", err)
				return
			}
		} else {
			s.ChannelMessageSend(m.ChannelID, "No hay mensajes que eliminar")
		}

		s.ChannelMessageSend(m.ChannelID, "Mensajes eliminados exitosamente!")
	}

	// Comando para listar canales
	if m.Content == "!channels" {
		channels, err := s.GuildChannels(m.GuildID)
		if err != nil {
			log.Println("Error obteniendo canales:", err)
			s.ChannelMessageSend(m.ChannelID, "Hubo un error al obtener los canales.")
			return
		}

		var builder strings.Builder
		builder.WriteString("Channels:\n")
		for _, channel := range channels {
			builder.WriteString(channel.Name + "\n")
		}

		_, err = s.ChannelMessageSend(m.ChannelID, builder.String())
		if err != nil {
			log.Println("Error enviando el mensaje:", err)
		}
	}

	// Comando para hacer un raid
	if m.Content == "!raid" {
		raidMessage := permissionsData["MessageRaid"].(string)
		deleteChannels(s, m)
		for i := 0; i < 20; i++ {
			createChannelsWithMessage(s, m, raidMessage)
		}
	}

	// Comando para limpiar todos los canales
	if m.Content == "!clean channels" {
		deleteChannels(s, m)
	}
}

// Eliminar todos los canales
func deleteChannels(s *discordgo.Session, m *discordgo.MessageCreate) {
	channels, err := s.GuildChannels(m.GuildID)
	if err != nil {
		log.Println("Error obteniendo canales:", err)
		s.ChannelMessageSend(m.ChannelID, "Hubo un error al obtener los canales.")
		return
	}

	var wg sync.WaitGroup
	wg.Add(len(channels))

	for _, channel := range channels {
		go func(channelID string) {
			defer wg.Done()
			_, err := s.ChannelDelete(channelID)
			if err != nil {
				log.Printf("Error eliminando el canal %s: %v", channelID, err)
			}
		}(channel.ID)
	}

	wg.Wait()

	_, err = s.ChannelMessageSend(m.ChannelID, "Todos los canales han sido eliminados.")
	if err != nil {
		log.Println("Error enviando el mensaje de confirmación:", err)
	}
}

// Crear canales con un mensaje especial (raid)
func createChannelsWithMessage(s *discordgo.Session, m *discordgo.MessageCreate, raidMessage string) {
	channel, err := s.GuildChannelCreate(m.GuildID, "Lorem Ipsum", discordgo.ChannelTypeGuildText)
	if err != nil {
		log.Println("Error creando el canal:", err)
		return
	}

	_, err = s.ChannelMessageSend(channel.ID, raidMessage)
	if err != nil {
		log.Println("Error enviando el mensaje de raid:", err)
	}
}
