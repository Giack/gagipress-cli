package prompts

import (
	"fmt"
	"strings"
)

// BookNiche represents different book categories
type BookNiche string

const (
	ChildrenBooks BookNiche = "children"
	Puzzles       BookNiche = "puzzles"
	Savings       BookNiche = "savings"
	DialectPuzzles BookNiche = "dialect_puzzles"
)

// IdeaPromptTemplate generates a prompt for content idea generation
func IdeaPromptTemplate(bookTitle, genre, targetAudience string, niche BookNiche, count int) string {
	baseContext := fmt.Sprintf(`Sei un esperto di social media marketing per libri self-published su Amazon KDP.

Il libro: "%s"
Genere: %s
Target: %s

Devi generare %d idee creative per contenuti TikTok/Instagram Reels che promuovano questo libro.`, bookTitle, genre, targetAudience, count)

	var nicheGuidelines string

	switch niche {
	case ChildrenBooks:
		nicheGuidelines = `
LINEE GUIDA per libri per bambini:
- Mostra momenti divertenti di lettura con i bambini
- Behind-the-scenes della creazione delle illustrazioni
- Consigli educativi per genitori
- Tutorial creativi ispirati al libro
- Storie animate delle pagine del libro
- Testimonianze di genitori e bambini`

	case Puzzles:
		nicheGuidelines = `
LINEE GUIDA per libri di enigmistica:
- Sfide e quiz interattivi dal libro
- Time-lapse di risoluzione enigmi
- Curiosità e trucchi per enigmisti
- Confronti "prima vs dopo" della mente
- Mini-sfide con premio (engagement)
- Spiegazione di enigmi particolarmente difficili`

	case DialectPuzzles:
		nicheGuidelines = `
LINEE GUIDA per enigmistica in dialetto milanese:
- Parole milanesi dimenticate con spiegazioni divertenti
- Confronto dialetto vs italiano standard
- Quiz su modi di dire milanesi
- Storielle brevi in dialetto
- Nostalgia e tradizioni milanesi
- Coinvolgimento community milanese`

	case Savings:
		nicheGuidelines = `
LINEE GUIDA per libri sul risparmio:
- Tips pratici di risparmio giornaliero
- Testimonianze di successo
- Sfide di risparmio da provare
- Errori comuni da evitare
- Trucchi psicologici per risparmiare
- Confronto spesa prima/dopo consigli del libro`
	}

	categories := `
CATEGORIE di contenuto (distribuisci equamente):
1. EDUCATIONAL: insegna qualcosa di utile
2. ENTERTAINMENT: diverte e intrattiene
3. BTS (Behind-The-Scenes): mostra il processo creativo
4. UGC (User Generated Content): coinvolge gli utenti
5. TREND: cavalca trend attuali di TikTok/Instagram

Per ogni idea, fornisci:
1. Tipo (educational/entertainment/bts/ugc/trend)
2. Titolo accattivante (max 10 parole)
3. Descrizione breve (2-3 frasi)
4. Hook iniziale suggerito
5. CTA finale suggerito
6. Punteggio rilevanza 0-100

Formato risposta (JSON array):
[
  {
    "type": "educational",
    "title": "Titolo idea",
    "description": "Descrizione dettagliata dell'idea",
    "hook": "Hook iniziale per catturare attenzione",
    "cta": "Call-to-action finale",
    "relevance_score": 85
  },
  ...
]`

	return baseContext + "\n" + nicheGuidelines + "\n" + categories
}

// ScriptPromptTemplate generates a prompt for script generation from an idea.
// amazonURL is the direct Amazon product link (e.g. https://www.amazon.it/dp/B0XXXXX).
// Pass an empty string if the book has no ASIN set.
func ScriptPromptTemplate(idea, bookTitle, platform, amazonURL string) string {
	platformSpecs := ""

	if platform == "tiktok" {
		platformSpecs = `
SPECIFICHE TIKTOK:
- Durata: 15-60 secondi
- Hook: primi 3 secondi CRITICI
- Ritmo: veloce, dinamico
- Formato: verticale 9:16
- Trend: usa musiche popolari
- Hashtag: 3-5 rilevanti + 2-3 di nicchia`
	} else if platform == "instagram" {
		platformSpecs = `
SPECIFICHE INSTAGRAM REELS:
- Durata: 15-90 secondi
- Hook: primi 3 secondi CRITICI
- Ritmo: medio-veloce
- Formato: verticale 9:16
- Audio: trending o originale
- Hashtag: 5-10 misti (popolari + nicchia)`
	}

	return fmt.Sprintf(`Sei un copywriter esperto di TikTok e Instagram Reels.

Idea da trasformare in script:
"%s"

Libro promosso: "%s"
Platform: %s

%s

Crea uno script completo strutturato così:

**HOOK (3-5 secondi)**
La frase/domanda che ferma lo scroll. Deve essere:
- Provocatoria o curiosa
- Relazionabile al target
- Chiara e diretta

**CONTENUTO PRINCIPALE (25-45 secondi)**
- Sviluppa l'idea in 3-5 punti chiave
- Linguaggio semplice e diretto
- Usa "tu" per parlare direttamente al viewer
- Include dettagli specifici e concreti

**CTA (5-10 secondi)**
- Invito all'azione chiaro
- Perché dovrebbero comprare il libro
- Link diretto: %s
- Menziona sia "link in bio" sia il link Amazon diretto

**EXTRA**
- 5-8 hashtag strategici
- Suggerimento musica/audio trending
- Note per il montaggio video

Formato risposta (JSON):
{
  "hook": "Hook text qui",
  "main_content": "Contenuto principale qui (separato in paragrafi)",
  "cta": "CTA text qui",
  "hashtags": ["#tag1", "#tag2", ...],
  "music_suggestion": "Nome traccia/audio trending",
  "video_notes": "Note per editing e montaggio",
  "estimated_length": 45
}`, idea, bookTitle, platform, platformSpecs, amazonURL)
}

// CalculateRelevanceScore calculates a relevance score for an idea
func CalculateRelevanceScore(ideaType, bookGenre string, hasBookReference bool, trendAlignment int) int {
	score := 50 // base score

	// Type alignment with genre
	typeScores := map[string]map[string]int{
		"children": {
			"educational":   +20,
			"entertainment": +15,
			"bts":           +10,
			"ugc":           +15,
			"trend":         +10,
		},
		"puzzles": {
			"educational":   +15,
			"entertainment": +20,
			"bts":           +5,
			"ugc":           +20,
			"trend":         +15,
		},
		"savings": {
			"educational":   +25,
			"entertainment": +10,
			"bts":           +10,
			"ugc":           +15,
			"trend":         +10,
		},
	}

	if genreScores, ok := typeScores[strings.ToLower(bookGenre)]; ok {
		if typeScore, ok := genreScores[ideaType]; ok {
			score += typeScore
		}
	}

	// Book reference bonus
	if hasBookReference {
		score += 10
	}

	// Trend alignment (0-20)
	score += trendAlignment

	// Cap at 100
	if score > 100 {
		score = 100
	}

	return score
}
