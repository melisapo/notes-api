package moderation

import (
	"strings"
	"unicode"
)

type Result struct {
	Approved bool
	Score    float64
	Reason   string
}

var blacklist = []string{
	// español
	"matar", "muerte a", "odio a", "imbécil", "idiota", "estúpido", "estúpida",
	"maldito", "maldita", "inútil", "basura", "asco", "golpear", "suicid",
	// english
	"kill", "hate", "stupid", "idiot", "worthless", "garbage", "die",
}

var positiveSet = map[string]bool{
	// español
	"bien": true, "bueno": true, "buena": true, "excelente": true,
	"amor": true, "feliz": true, "alegria": true, "esperanza": true,
	"bonito": true, "bonita": true, "genial": true, "increible": true,
	"maravilloso": true, "exito": true, "gracias": true, "paz": true,
	"sonrisa": true, "fuerza": true, "animo": true, "juntos": true,
	// english
	"good": true, "great": true, "love": true, "happy": true,
	"joy": true, "hope": true, "beautiful": true, "amazing": true,
	"wonderful": true, "success": true, "thanks": true, "peace": true,
	"smile": true, "together": true,
}

var negativeSet = map[string]bool{
	// español
	"malo": true, "mala": true, "terrible": true, "horrible": true,
	"pesimo": true, "pesima": true, "triste": true, "fracaso": true,
	"fracasado": true, "feo": true, "fea": true, "nunca": true,
	"imposible": true, "aburrido": true, "mentira": true, "mentiroso": true,
	"falso": true, "falsa": true, "peor": true, "fatal": true, "desastre": true,
	// english
	"bad": true, "awful": true, "sad": true, "ugly": true,
	"boring": true, "worse": true, "worst": true, "failure": true,
	"useless": true, "disaster": true, "lie": true, "fake": true,
}

func normalize(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		switch r {
		case 'á':
			b.WriteRune('a')
		case 'é':
			b.WriteRune('e')
		case 'í':
			b.WriteRune('i')
		case 'ó':
			b.WriteRune('o')
		case 'ú', 'ü':
			b.WriteRune('u')
		case 'ñ':
			b.WriteRune('n')
		default:
			if !unicode.IsPunct(r) {
				b.WriteRune(r)
			}
		}
	}
	return b.String()
}

func Moderate(content string) Result {
	text := normalize(content)
	words := strings.Fields(text)

	if len(words) == 0 {
		return Result{false, 0, "empty content"}
	}

	for _, banned := range blacklist {
		if strings.Contains(text, banned) {
			return Result{false, 0, "innapropiate content"}
		}
	}

	var pos, neg int
	for _, word := range words {
		if positiveSet[word] {
			pos++
		}
		if negativeSet[word] {
			neg++
		}
	}

	score := 0.5 + float64(pos)*0.1 - float64(neg)*0.1
	if score > 1.0 {
		score = 1.0
	}
	if score < 0.0 {
		score = 0.0
	}

	switch {
	case score >= 0.5:
		return Result{true, score, "approved"}
	case score >= 0.3:
		return Result{true, score, "negative words detected"}
	default:
		return Result{false, score, "too negative"}
	}
}
