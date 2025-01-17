package discord

import (
	"github.com/bwmarrin/discordgo"
	"strings"
	"time"
)

func ConstructInfoEmbed(participants []string, giveawayHours string) *discordgo.MessageEmbed {
	info := "**Ten bot organizuje giveaway kodów na serwery Diamond.**\n" +
		"**Każdy kod przedłuża serwer o 7 dni.**\n" +
		"Aby wziąć udział pomagaj innym użytkownikom. Jeżeli komuś pomożesz, to poproś tą osobę aby użyła komendy </thx:1107007500659728405> lub sam użyj komendy </thxme:1107007504308769020> - w ten sposób dostaniesz się do loterii. To jest nasza metoda na rozruszanie tego Discorda, tak, aby każdy mógł liczyć na pomoc. Każde podziękowanie to jeden los, więc warto pomagać!\n\n" +
		"**Sponsorem tego bota jest https://craftserve.pl/ - hosting serwerów Minecraft.**\n\n" +
		"Pomoc musi odbywać się na tym serwerze na tekstowych kanałach publicznych.\n\n" +
		"Uczestnicy: " + strings.Join(participants, ", ") + "\n\nNagrody rozdajemy o " + giveawayHours + ", Powodzenia!"
	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			URL:     "https://craftserve.pl",
			Name:    "Informacje o Giveawayu",
			IconURL: "https://cdn.discordapp.com/avatars/524308413719642118/c2a17b4479bfcc89d2b7e64e6ae15ebe.webp",
		},
		Description: info,
		Color:       0x234d20,
		Timestamp:   time.Now().Format(time.RFC3339),
	}
	return embed
}

func ConstructThxEmbed(participants []string, giveawayHours, participantId, confirmerId, state string) *discordgo.MessageEmbed {
	embed := ConstructInfoEmbed(participants, giveawayHours)
	embed.Fields = []*discordgo.MessageEmbedField{}
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Dodany", Value: "<@" + participantId + ">", Inline: true})

	var status string
	switch state {
	case "wait":
		status = "Oczekiwanie"
		break
	case "confirm":
		status = "Potwierdzono"
		if confirmerId != "" {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Potwierdzający", Value: "<@" + confirmerId + ">", Inline: true})
		}
		break
	case "reject":
		status = "Odrzucono"
		if confirmerId != "" {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Odrzucający", Value: "<@" + confirmerId + ">", Inline: true})
		}
		break
	}
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Status", Value: status, Inline: true})

	return embed
}

func ConstructThxNotificationEmbed(guildId, thxChannelId, thxMessageId, participantId, confirmerId, state string) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Timestamp: time.Now().Format(time.RFC3339),
		Author: &discordgo.MessageEmbedAuthor{
			URL:     "https://craftserve.pl",
			Name:    "Nowe podziękowanie",
			IconURL: "https://cdn.discordapp.com/avatars/524308413719642118/c2a17b4479bfcc89d2b7e64e6ae15ebe.webp",
		},
		Color: 0x234d20,
	}
	embed.Fields = []*discordgo.MessageEmbedField{}
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Dla", Value: "<@" + participantId + ">", Inline: true})
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Kanał", Value: "<#" + thxChannelId + ">", Inline: true})

	switch state {
	case "wait":
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Status", Value: "Oczekiwanie", Inline: true})
		break
	case "confirm":
		if confirmerId != "" {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Status", Value: "Potwierdzono przez <@" + confirmerId + ">", Inline: true})
		}
		break
	case "reject":
		if confirmerId == "" {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Status", Value: "Odrzucono", Inline: true})
		} else {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Status", Value: "Odrzucono przez <@" + confirmerId + ">", Inline: true})
		}
		break
	}
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Link", Value: "https://discordapp.com/channels/" + guildId + "/" + thxChannelId + "/" + thxMessageId, Inline: false})

	return embed
}

func ConstructWinnerEmbed(code string) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			URL:     "https://craftserve.pl",
			Name:    "Wygrałeś kod na serwer diamond!",
			IconURL: "https://cdn.discordapp.com/avatars/524308413719642118/c2a17b4479bfcc89d2b7e64e6ae15ebe.webp",
		},
		Description: "Gratulacje! W loterii wygrałeś darmowy kod na serwer w Craftserve! Możesz go użyć w zakładce *Płatności* pod przyciskiem *Zrealizuj kod podarunkowy*. Kod jest ważny około rok.",
		Color:       0x234d20,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name: "KOD", Value: code,
			},
		},
	}
	return embed
}

func ConstructMessageWinnerEmbed(codes []string) *discordgo.MessageEmbed {
	var description string
	var author string
	var name string
	if len(codes) > 1 {
		description = "Gratulacje! W loterii wygrałeś darmowe kody na serwer w Craftserve! Możesz je użyć w zakładce *Płatności* pod przyciskiem *Zrealizuj kod podarunkowy*. Kody są ważne około rok."
		author = "Wygrałeś kody na serwer diamond!"
		name = "KODY"
	} else {
		description = "Gratulacje! W loterii wygrałeś darmowy kod na serwer w Craftserve! Możesz go użyć w zakładce *Płatności* pod przyciskiem *Zrealizuj kod podarunkowy*. Kod jest ważny około rok."
		author = "Wygrałeś kod na serwer diamond!"
		name = "KOD"
	}
	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			URL:     "https://craftserve.pl",
			Name:    author,
			IconURL: "https://cdn.discordapp.com/avatars/524308413719642118/c2a17b4479bfcc89d2b7e64e6ae15ebe.webp",
		},
		Color:       0x234d20,
		Description: description,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name: name, Value: strings.Join(codes, "\n"),
			},
		},
	}
	return embed
}

func ConstructChannelWinnerEmbed(username string) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			URL:     "https://craftserve.pl",
			Name:    "Wyniki giveaway!",
			IconURL: "https://cdn.discordapp.com/avatars/524308413719642118/c2a17b4479bfcc89d2b7e64e6ae15ebe.webp",
		},
		Color:       0x234d20,
		Description: username + " wygrał kod. Gratulacje!",
	}
	return embed
}

func ConstructChannelMessageWinnerEmbed(usernames []string) *discordgo.MessageEmbed {
	var description string
	if len(usernames) > 1 {
		description = "W loterii za aktywność wygraliście kody na serwer w Craftserve!"
		description += "\n\n" + strings.Join(usernames, "\n")
	} else {
		description = "W loterii za aktywność " + usernames[0] + " wygrał kod na serwer w Craftserve. Gratulacje!"
	}
	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			URL:     "https://craftserve.pl",
			Name:    "Wyniki giveaway!",
			IconURL: "https://cdn.discordapp.com/avatars/524308413719642118/c2a17b4479bfcc89d2b7e64e6ae15ebe.webp",
		},
		Color:       0x234d20,
		Description: description,
	}
	return embed
}

func ConstructResendEmbed(codes []string) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			URL:     "https://craftserve.pl",
			Name:    "Twoje ostatnie wygrane kody",
			IconURL: "https://cdn.discordapp.com/avatars/524308413719642118/c2a17b4479bfcc89d2b7e64e6ae15ebe.webp",
		},
		Description: strings.Join(codes, "\n"),
		Color:       0x234d20,
		Timestamp:   time.Now().Format(time.RFC3339),
	}
	return embed
}
