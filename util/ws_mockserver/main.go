package main

import (
	"fmt"
	"log"
	"main/util/ws_mockserver/lib"
	"net/http"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	appStyle = lipgloss.NewStyle().Padding(1, 2)

	connectionContainer = lipgloss.NewStyle().
				Margin(1, 3, 0)

	connectionStatusLabel = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#F0F0FF")).
				Background(lipgloss.Color("#000000")).
				Padding(0, 1)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F0F0FF")).
			Background(lipgloss.Color("#9146FF")).
			Padding(0, 1)
)

type model struct {
	eventChan   chan<- string
	statusChan  <-chan string
	showFormFor string
	statusText  string
	list        list.Model
	form        *huh.Form
}

func newModel(eventChan chan<- string, statusChan <-chan string) model {
	m := model{
		eventChan:  eventChan,
		statusChan: statusChan,
		statusText: "Not connected",
	}

	eventList := list.New(lib.EventItems, newItemDelegate(), 0, 0)
	eventList.Title = "Twitch Event Testserverino"
	eventList.Styles.Title = titleStyle

	m.list = eventList
	return m
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	if len(m.statusChan) >= 1 {
		m.statusText = <-m.statusChan
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := appStyle.GetFrameSize()

		_, headerHeight := lipgloss.Size(connectionContainer.Render(lipgloss.JoinHorizontal(
			lipgloss.Center,
			connectionStatusLabel.Render("Connection status:"), titleStyle.Render(m.statusText),
		)))

		m.list.SetSize(msg.Width-(h), msg.Height-(v+headerHeight))
	case switchPageMsg:
		m.showFormFor = msg.eventName
		m.form = msg.makeForm()
		return m, m.form.Init()
	}

	if m.showFormFor == "" {
		// This will also call our delegate's update function.
		newListModel, cmd := m.list.Update(msg)
		m.list = newListModel
		cmds = append(cmds, cmd)
	} else {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if msg.Type == tea.KeyEsc {
				m.showFormFor = ""
				return m, nil
			}
		}

		if m.form != nil {
			newFormModel, cmd := m.form.Update(msg)
			if f, ok := newFormModel.(*huh.Form); ok {
				m.form = f
			}
			cmds = append(cmds, cmd)

			if m.form.State == huh.StateCompleted {
				selectedItem := m.list.SelectedItem().(lib.EventItem)
				eventMsg := selectedItem.MakePayload(m.form)
				m.eventChan <- eventMsg

				m.showFormFor = ""
				m.form = nil
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.showFormFor != "" {
		return fmt.Sprintf(
			"%s\n%s",
			titleStyle.Render(fmt.Sprintf("Event prep for %s", m.showFormFor)),
			m.form.View(),
		)
	}

	return fmt.Sprintf(
		"%s\n%s",
		connectionContainer.Render(lipgloss.JoinHorizontal(
			lipgloss.Center,
			connectionStatusLabel.Render("Connection status:"), titleStyle.Render(m.statusText),
		)),
		appStyle.Render(m.list.View()),
	)
}

var chooseKeyBind = key.NewBinding(
	key.WithKeys("enter"),
	key.WithHelp("enter", "choose"),
)

type switchPageMsg struct {
	eventName string
	makeForm  func() *huh.Form
}

func switchFormPage(e lib.EventItem) tea.Cmd {
	return func() tea.Msg {
		return switchPageMsg{e.TwitchIdent(), e.MakeForm}
	}
}

func newItemDelegate() list.DefaultDelegate {
	d := list.NewDefaultDelegate()

	d.UpdateFunc = func(msg tea.Msg, m *list.Model) tea.Cmd {
		if _, ok := m.SelectedItem().(lib.EventItem); !ok {
			return nil
		}

		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, chooseKeyBind):
				return switchFormPage(m.SelectedItem().(lib.EventItem))
			}
		}

		return nil
	}

	help := []key.Binding{chooseKeyBind}

	d.ShortHelpFunc = func() []key.Binding {
		return help
	}

	d.FullHelpFunc = func() [][]key.Binding {
		return [][]key.Binding{help}
	}

	return d
}

func main() {
	eventChannel := make(chan string)
	statusChan := make(chan string, 1)
	m := newModel(eventChannel, statusChan)

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		statusChan <- "Connected!!"

		if err := conn.WriteMessage(websocket.TextMessage, []byte(`{
  "metadata": {
    "message_id": "96a3f3b5-5dec-4eed-908e-e11ee657416c",
    "message_type": "session_welcome",
    "message_timestamp": "2023-07-19T14:56:51.634234626Z"
  },
  "payload": {
    "session": {
      "id": "AQoQILE98gtqShGmLD7AM6yJThAB",
      "status": "connected",
      "connected_at": "2023-07-19T14:56:51.616329898Z",
      "keepalive_timeout_seconds": 10,
      "reconnect_url": null
    }
  }
}`)); err != nil {
			panic(err)
		}

		go func() {
			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					panic(err)
				}
			}
		}()
		go func() {
			for {
				strMsg := <-eventChannel
				if err := conn.WriteMessage(websocket.TextMessage, []byte(strMsg)); err != nil {
					panic(err)
				}
				statusChan <- fmt.Sprintf("Send event at %s", time.Now().Format("15:04:05"))
			}
		}()
	})
	go func() {
		err := http.ListenAndServe(":8190", nil)
		if err != nil {
			fmt.Println("Error starting WS server:", err)
			os.Exit(1)
		}
	}()

	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
