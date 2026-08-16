// Command slack-setup provisions the NAEOS Slack workspace: creates channels
// from launch/slack-server/blueprint.md and optionally posts paste-ready
// messages. Idempotent — safe to re-run.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var apiBase = "https://slack.com/api/"

// channelPlan describes one channel to create.
type channelPlan struct {
	name  string
	topic string
}

// blueprintChannels mirrors launch/slack-server/blueprint.md §1.
var blueprintChannels = []channelPlan{
	{name: "announcements", topic: "Releases, Product Hunt launch, pre-launch posts"},
	{name: "general", topic: "Main discussion"},
	{name: "introductions", topic: "New members say hi"},
	{name: "showcase", topic: "Share projects built with NAEOS"},
	{name: "software-architecture", topic: "Design discussions, patterns, NEIR model"},
	{name: "ai-engineering", topic: "AI-assisted engineering topics"},
	{name: "ai-agent", topic: "AI agent workflows, MCP server, instruction sets"},
	{name: "spec-language", topic: "Spec Language v2 ($ref, $fn, $if, migrations)"},
	{name: "code-generation", topic: "Multi-language generation (Go/TS/Py/Java/Rust)"},
	{name: "plugins", topic: "WASM plugin SDK, official example plugins"},
	{name: "help", topic: "Questions + answers"},
	{name: "off-topic", topic: "Casual chat"},
	{name: "launch-upvotes", topic: "PH link + how to support on launch day"},
	{name: "launch-day", topic: "Countdown, go-live watch, post-launch recap"},
}

// apiCall performs a Slack Web API call and decodes the envelope.
func apiCall(ctx context.Context, token, method string, params url.Values, out any) error {
	params.Set("token", token)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiBase+method, strings.NewReader(params.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return err
	}

	var env struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		return fmt.Errorf("%s: decode envelope: %w", method, err)
	}
	if !env.OK {
		return fmt.Errorf("%s: %s", method, env.Error)
	}
	if out != nil {
		if err := json.Unmarshal(body, out); err != nil {
			return fmt.Errorf("%s: decode payload: %w", method, err)
		}
	}
	return nil
}

// findChannelID returns the ID of a public channel by name, or "" if absent.
func findChannelID(ctx context.Context, token, name string) (string, error) {
	var out struct {
		Channels []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"channels"`
	}
	params := url.Values{"limit": {"1000"}, "exclude_archived": {"true"}}
	if err := apiCall(ctx, token, "conversations.list", params, &out); err != nil {
		return "", err
	}
	for _, ch := range out.Channels {
		if ch.Name == name {
			return ch.ID, nil
		}
	}
	return "", nil
}

// createChannel creates a public channel (idempotent: returns existing ID).
func createChannel(ctx context.Context, token string, p channelPlan) (string, error) {
	if id, err := findChannelID(ctx, token, p.name); err != nil {
		return "", err
	} else if id != "" {
		return id, nil
	}
	var out struct {
		Channel struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"channel"`
	}
	params := url.Values{"name": {p.name}}
	if err := apiCall(ctx, token, "conversations.create", params, &out); err != nil {
		return "", err
	}
	log.Printf("created #%s (%s)", out.Channel.Name, out.Channel.ID)
	return out.Channel.ID, nil
}

// postMessage posts text to a channel.
func postMessage(ctx context.Context, token, channelID, text string) (string, error) {
	var out struct {
		Ts string `json:"ts"`
	}
	params := url.Values{"channel": {channelID}, "text": {text}}
	if err := apiCall(ctx, token, "chat.postMessage", params, &out); err != nil {
		return "", err
	}
	return out.Ts, nil
}

const welcomeMessage = `Welcome to the NAEOS Community 👋

NAEOS is a declarative engineering platform: describe your system once, and it builds, validates, and evolves real software — for humans and AI.

*Specify once. Build anywhere.*

*Start here*
1. Say hi in <#introductions>.
2. Try NAEOS in 30 seconds:
` + "```bash\ncurl -fsSL https://naeos.dev/install.sh | sh\nnaeos create\ncd my-app && naeos run --input-file spec.yaml\n```" + `
3. Get help in <#help>, share what you build in <#showcase>.

*Useful links*
- Docs: https://docs.naeos.dev
- GitHub: https://github.com/NAEOS-foundation/naeos
- Website: https://naeos.dev
- Whitepaper: https://naeos.dev/whitepaper`

const rulesMessage = `*Community Rules*
1. *Be excellent to each other.* Disagree with ideas, not people.
2. *Stay on topic.* Engineering topics in the engineering channels; casual in <#off-topic>.
3. *Search before asking.* Check docs first, then ask in <#help> with your spec/CLI output.
4. *No spam.* Self-promotion belongs in <#showcase> and must add value. No vote-brigading or paid promotion, ever.
5. *Respect the license.* NAEOS is Apache 2.0 — keep attribution where required.`

const preLaunchMessage = `*NAEOS is launching Tuesday 🚀*

On *Tuesday, 18 Aug* we launch NAEOS on Product Hunt.

NAEOS is a declarative engineering platform. You describe your system *once* in YAML/JSON — it builds an internal engineering model (NEIR) and generates validated code in Go, TypeScript, Python, Java, Rust, plus AI instruction sets for 6 tools and an MCP server.

*This week:*
- 👀 Community walkthrough of v3.1.0 (caching, profiling, architecture patterns)
- 🗳️ Roadmap voting — tell us what ships next
- 🏆 Become a Launch Champion — help us launch
- 🎉 Launch day — watch party + support thread in <#launch-upvotes>

Try NAEOS right now:
` + "```bash\ncurl -fsSL https://naeos.dev/install.sh | sh\nnaeos create\ncd my-app && naeos run --input-file spec.yaml\n```" + `
Links: https://naeos.dev · https://docs.naeos.dev · https://github.com/NAEOS-foundation/naeos`

func main() {
	flagChannels := flag.Bool("channels", false, "create channels from the blueprint")
	flagMessages := flag.Bool("messages", false, "post welcome/rules/pre-launch messages to announcements")
	flag.Parse()

	token := os.Getenv("NAEOS_SLACK_TOKEN")
	if token == "" {
		log.Fatal("NAEOS_SLACK_TOKEN is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if !*flagChannels && !*flagMessages {
		log.Fatal("specify at least one of -channels or -messages")
	}

	if *flagChannels {
		for _, p := range blueprintChannels {
			id, err := createChannel(ctx, token, p)
			if err != nil {
				log.Printf("channel %s: %v", p.name, err)
				continue
			}
			if p.topic != "" {
				if err := apiCall(ctx, token, "conversations.setTopic",
					url.Values{"channel": {id}, "topic": {p.topic}}, nil); err != nil {
					log.Printf("topic %s: %v", p.name, err)
				}
			}
		}
	}

	if *flagMessages {
		announceID, err := findChannelID(ctx, token, "announcements")
		if err != nil {
			log.Fatal("find announcements: ", err)
		}
		if announceID == "" {
			log.Fatal("announcements channel not found — run with -channels first")
		}
		msgs := []struct{ name, text string }{
			{"welcome", welcomeMessage},
			{"rules", rulesMessage},
			{"pre-launch", preLaunchMessage},
		}
		for _, m := range msgs {
			if ts, err := postMessage(ctx, token, announceID, m.text); err != nil {
				log.Printf("post %s: %v", m.name, err)
			} else {
				log.Printf("posted %s (ts=%s)", m.name, ts)
			}
		}
	}
}
