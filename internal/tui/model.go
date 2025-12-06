package tui

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/cedev-1/jellyfin-mustui/internal/config"
	"github.com/cedev-1/jellyfin-mustui/internal/jellyfin"
	"github.com/cedev-1/jellyfin-mustui/internal/player"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type sessionState int

const (
	stateLogin sessionState = iota
	stateLibraryList
	stateMusicPlayer
)

type panelFocus int

const (
	focusArtists panelFocus = iota
	focusTracks
)

type Model struct {
	cfg    *config.Config
	client *jellyfin.Client
	player *player.Player
	state  sessionState

	loginInputs []textinput.Model
	focusIndex  int
	err         error

	libraryList list.Model
	artistList  list.Model
	trackList   list.Model

	artists            []jellyfin.MusicItem
	currentArtist      *jellyfin.MusicItem
	albums             []jellyfin.MusicItem
	selectedAlbumIndex int
	tracks             []jellyfin.MusicItem

	panelFocus   panelFocus
	position     time.Duration
	duration     time.Duration
	isPlaying    bool
	isLoading    bool
	showHelp     bool
	currentTrack *player.Track
	progressBar  progress.Model

	width  int
	height int
}

type tickMsg time.Time
type progressMsg struct {
	position time.Duration
	duration time.Duration
}
type trackChangedMsg *player.Track
type stateChangedMsg player.State
type artistsLoadedMsg []jellyfin.MusicItem
type albumsLoadedMsg []jellyfin.MusicItem
type tracksLoadedMsg []jellyfin.MusicItem
type trackReadyMsg struct {
	track      *player.Track
	queueIndex int
	err        error
}
type errMsg error

func NewModel(cfg *config.Config, client *jellyfin.Client) Model {
	m := Model{
		cfg:        cfg,
		client:     client,
		player:     player.New(),
		state:      stateLogin,
		panelFocus: focusArtists,
	}

	m.libraryList = list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	m.artistList = list.New([]list.Item{}, musicDelegate{}, 0, 0)
	m.trackList = list.New([]list.Item{}, musicDelegate{}, 0, 0)
	m.artistList.Title = "Artists"
	m.trackList.Title = "Tracks"

	m.artistList.SetShowHelp(false)
	m.trackList.SetShowHelp(false)
	m.libraryList.SetShowHelp(false)

	if cfg.Token != "" && cfg.ServerURL != "" && cfg.UserID != "" {
		m.state = stateMusicPlayer
	} else {
		m.loginInputs = make([]textinput.Model, 3)
		var t textinput.Model
		for i := range m.loginInputs {
			t = textinput.New()
			t.CharLimit = 64

			switch i {
			case 0:
				t.Placeholder = "Server URL"
				t.Focus()
				t.PromptStyle = inputFocusedStyle
				t.TextStyle = inputFocusedStyle
				if cfg.ServerURL != "" {
					t.SetValue(cfg.ServerURL)
				}
			case 1:
				t.Placeholder = "Username"
				t.PromptStyle = inputBlurredStyle
				t.TextStyle = inputBlurredStyle
			case 2:
				t.Placeholder = "Password"
				t.EchoMode = textinput.EchoPassword
				t.EchoCharacter = '•'
				t.PromptStyle = inputBlurredStyle
				t.TextStyle = inputBlurredStyle
			}

			m.loginInputs[i] = t
		}
	}

	m.player.OnProgress = func(pos, dur time.Duration) {
	}
	m.player.OnTrackChange = func(track *player.Track) {
		m.currentTrack = track
	}
	m.player.OnStateChange = func(state player.State) {
		m.isPlaying = state == player.StatePlaying
	}

	m.progressBar = progress.New(progress.WithSolidFill(string(colorSubtext)))

	return m
}

func (m Model) Init() tea.Cmd {
	if err := m.player.Init(); err != nil {
		return func() tea.Msg { return errMsg(err) }
	}
	if m.state == stateMusicPlayer {
		return tea.Batch(m.loadArtists, m.tickCmd())
	}
	return textinput.Blink
}

func (m Model) tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m Model) loadArtists() tea.Msg {
	artists, err := m.client.GetArtists()
	if err != nil {
		return errMsg(err)
	}
	return artistsLoadedMsg(artists)
}

func (m Model) loadAlbums(artistID string) tea.Cmd {
	return func() tea.Msg {
		albums, err := m.client.GetAlbums(artistID)
		if err != nil {
			return errMsg(err)
		}
		return albumsLoadedMsg(albums)
	}
}

func (m Model) loadTracks(albumID string) tea.Cmd {
	return func() tea.Msg {
		tracks, err := m.client.GetTracks(albumID)
		if err != nil {
			return errMsg(err)
		}
		return tracksLoadedMsg(tracks)
	}
}

func (m Model) playTrackAsync(index int) tea.Cmd {
	return func() tea.Msg {
		var err error
		switch index {
		case -1:
			err = m.player.Next()
		case -2:
			err = m.player.Previous()
		default:
			err = m.player.PlayFromQueue(index)
		}
		if err != nil {
			return trackReadyMsg{track: nil, queueIndex: -1, err: err}
		}
		return trackReadyMsg{
			track:      m.player.GetCurrentTrack(),
			queueIndex: m.player.GetQueueIndex(),
			err:        nil,
		}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.player.Close()
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		listHeight := m.height - 10
		m.artistList.SetSize(m.width/3, listHeight)
		m.trackList.SetSize(m.width*2/3, listHeight)
		m.progressBar.Width = m.width - 30
		if m.progressBar.Width > 80 {
			m.progressBar.Width = 80
		}
	case tickMsg:
		var cmds []tea.Cmd
		if m.player.GetState() == player.StatePlaying {
			m.position = m.player.GetPosition()
			track := m.player.GetCurrentTrack()
			if track != nil {
				m.currentTrack = track
				m.duration = track.Duration
			}
			if m.duration > 0 {
				percent := float64(m.position) / float64(m.duration)
				cmds = append(cmds, m.progressBar.SetPercent(percent))
			}
		}
		cmds = append(cmds, m.tickCmd())
		return m, tea.Batch(cmds...)
	case progress.FrameMsg:
		progressModel, cmd := m.progressBar.Update(msg)
		m.progressBar = progressModel.(progress.Model)
		return m, cmd
	case trackReadyMsg:
		m.isLoading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.currentTrack = msg.track
		if msg.track != nil {
			m.duration = msg.track.Duration
		}
		if msg.queueIndex >= 0 {
			m.trackList.Select(msg.queueIndex)
		}
		m.isPlaying = true
		m.err = nil
	case errMsg:
		m.err = msg
	}

	switch m.state {
	case stateLogin:
		return m.updateLogin(msg)
	case stateLibraryList:
		return m.updateLibraryList(msg)
	case stateMusicPlayer:
		return m.updateMusicPlayer(msg)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	switch m.state {
	case stateLogin:
		return m.viewLogin()
	case stateLibraryList:
		return m.libraryList.View()
	case stateMusicPlayer:
		return m.viewMusicPlayer()
	}
	return "Unknown state"
}

func (m Model) updateLogin(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case *jellyfin.AuthResponse:
		m.state = stateMusicPlayer
		return m, tea.Batch(m.loadArtists, m.tickCmd())

	case error:
		m.err = msg
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			if s == "enter" && m.focusIndex == len(m.loginInputs) {
				return m, m.performLogin
			}

			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.loginInputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.loginInputs)
			}

			cmds := make([]tea.Cmd, len(m.loginInputs))
			for i := 0; i <= len(m.loginInputs)-1; i++ {
				if i == m.focusIndex {
					cmds[i] = m.loginInputs[i].Focus()
					m.loginInputs[i].PromptStyle = inputFocusedStyle
					m.loginInputs[i].TextStyle = inputFocusedStyle
				} else {
					m.loginInputs[i].Blur()
					m.loginInputs[i].PromptStyle = inputBlurredStyle
					m.loginInputs[i].TextStyle = inputBlurredStyle
				}
			}
			return m, tea.Batch(cmds...)
		}
	}

	cmd := m.updateInputs(msg)
	return m, cmd
}

func (m *Model) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.loginInputs))
	for i := range m.loginInputs {
		m.loginInputs[i], cmds[i] = m.loginInputs[i].Update(msg)
	}
	return tea.Batch(cmds...)
}

func (m Model) viewLogin() string {
	title := titleStyle.Render("JELLYFIN-MUSTUI")

	inputs := make([]string, len(m.loginInputs))
	for i := range m.loginInputs {
		inputs[i] = m.loginInputs[i].View()
	}

	btn := buttonStyle.Render("Login")
	if m.focusIndex == len(m.loginInputs) {
		btn = activeButtonStyle.Render("Login")
	}

	content := lipgloss.JoinVertical(lipgloss.Center,
		title,
		"\n",
		strings.Join(inputs, "\n\n"),
		"\n",
		btn,
	)

	if m.err != nil {
		content = lipgloss.JoinVertical(lipgloss.Center, content, errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
	}

	box := loginBoxStyle.Render(content)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func (m Model) performLogin() tea.Msg {
	url := m.loginInputs[0].Value()
	user := m.loginInputs[1].Value()
	pass := m.loginInputs[2].Value()

	m.client.ServerURL = url
	resp, err := m.client.Authenticate(user, pass)
	if err != nil {
		return err
	}

	m.cfg.ServerURL = url
	m.cfg.Username = user
	m.cfg.Token = resp.AccessToken
	m.cfg.UserID = resp.User.ID
	if err := config.SaveConfig(m.cfg); err != nil {
		return err
	}

	return resp
}

func (m Model) updateLibraryList(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.libraryList, cmd = m.libraryList.Update(msg)
	return m, cmd
}

func (m Model) updateMusicPlayer(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case artistsLoadedMsg:
		m.artists = msg
		items := make([]list.Item, len(msg))
		for i, a := range msg {
			items[i] = musicItem{a}
		}
		m.artistList = list.New(items, musicDelegate{}, m.width/3, m.height-10)
		m.artistList.Title = "Artists"
		m.artistList.Styles.Title = listTitleStyle
		m.artistList.SetShowHelp(false)

	case albumsLoadedMsg:
		m.albums = msg
		m.selectedAlbumIndex = 0
		if len(msg) > 0 {
			return m, m.loadTracks(msg[0].ID)
		}

	case tracksLoadedMsg:
		m.tracks = msg

		items := make([]list.Item, len(msg))
		for i, t := range msg {
			items[i] = trackItem{MusicItem: t, queueIndex: i}
		}

		m.trackList = list.New(items, trackDelegate{}, m.width*2/3, m.height-10)
		if len(m.albums) > 0 && m.selectedAlbumIndex < len(m.albums) {
			m.trackList.Title = m.albums[m.selectedAlbumIndex].Name
		} else {
			m.trackList.Title = "Tracks"
		}
		m.trackList.Styles.Title = listTitleStyle
		m.trackList.SetShowHelp(false)

		queue := make([]player.Track, len(msg))
		for i, t := range msg {
			queue[i] = player.Track{
				ID:       t.ID,
				Name:     t.Name,
				Artist:   t.AlbumArtist,
				Album:    t.Album,
				Duration: time.Duration(t.RunTimeTicks/10000000) * time.Second,
				URL:      m.client.GetAudioStreamURL(t.ID),
			}
		}
		m.player.SetQueue(queue)

	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			if m.panelFocus == focusArtists {
				m.panelFocus = focusTracks
			} else {
				m.panelFocus = focusArtists
			}
		case "h", "left":
			if m.panelFocus == focusTracks && len(m.albums) > 0 {
				m.selectedAlbumIndex--
				if m.selectedAlbumIndex < 0 {
					m.selectedAlbumIndex = len(m.albums) - 1
				}
				return m, m.loadTracks(m.albums[m.selectedAlbumIndex].ID)
			}
		case "l", "right":
			if m.panelFocus == focusTracks && len(m.albums) > 0 {
				m.selectedAlbumIndex++
				if m.selectedAlbumIndex >= len(m.albums) {
					m.selectedAlbumIndex = 0
				}
				return m, m.loadTracks(m.albums[m.selectedAlbumIndex].ID)
			}
		case " ":
			m.player.TogglePause()
			m.isPlaying = m.player.GetState() == player.StatePlaying
		case "n":
			m.isLoading = true
			return m, m.playTrackAsync(-1)
		case "p":
			m.isLoading = true
			return m, m.playTrackAsync(-2)
		case "enter":
			if m.panelFocus == focusArtists {
				if item, ok := m.artistList.SelectedItem().(musicItem); ok {
					m.currentArtist = &item.MusicItem
					m.panelFocus = focusTracks
					return m, m.loadAlbums(item.ID)
				}
			} else {
				if item, ok := m.trackList.SelectedItem().(trackItem); ok {
					m.isLoading = true
					return m, m.playTrackAsync(item.queueIndex)
				}
			}
		case "q":
			m.player.Close()
			return m, tea.Quit
		case "?":
			m.showHelp = !m.showHelp
			return m, nil
		case "esc":
			if m.showHelp {
				m.showHelp = false
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	if m.panelFocus == focusArtists {
		m.artistList, cmd = m.artistList.Update(msg)
	} else {
		m.trackList, cmd = m.trackList.Update(msg)
	}
	return m, cmd
}

func (m Model) viewMusicPlayer() string {
	panelHeight := m.height - 12
	if panelHeight < 5 {
		panelHeight = 5
	}
	header := titleStyle.Render("♪ JELLYFIN-MUSTUI")
	headerCentered := lipgloss.PlaceHorizontal(m.width, lipgloss.Center, header)

	artistWidth := m.width/3 - 2
	trackWidth := m.width*2/3 - 4

	listHeight := panelHeight - 2
	if listHeight < 3 {
		listHeight = 3
	}
	m.artistList.SetSize(artistWidth-2, listHeight)
	m.trackList.SetSize(trackWidth-2, listHeight)

	artistStyle := panelStyle.Width(artistWidth).Height(panelHeight)
	if m.panelFocus == focusArtists {
		artistStyle = activePanelStyle.Width(artistWidth).Height(panelHeight)
	}
	artistPanel := artistStyle.Render(m.artistList.View())

	albumIndicator := ""
	if len(m.albums) > 0 {
		albumIndicator = fmt.Sprintf("◀ Album %d/%d ▶", m.selectedAlbumIndex+1, len(m.albums))
	}

	trackContent := m.trackList.View()
	if albumIndicator != "" {
		trackContent = albumHeaderStyle.Render(albumIndicator) + "\n" + trackContent
	}

	trackStyle := panelStyle.Width(trackWidth).Height(panelHeight)
	if m.panelFocus == focusTracks {
		trackStyle = activePanelStyle.Width(trackWidth).Height(panelHeight)
	}
	trackPanel := trackStyle.Render(trackContent)

	panels := lipgloss.JoinHorizontal(lipgloss.Top, artistPanel, trackPanel)
	panelsCentered := lipgloss.PlaceHorizontal(m.width, lipgloss.Center, panels)

	nowPlaying := m.renderNowPlaying()
	nowPlayingCentered := lipgloss.PlaceHorizontal(m.width, lipgloss.Center, nowPlaying)

	var errorCentered string
	if m.err != nil {
		errorCentered = lipgloss.PlaceHorizontal(m.width, lipgloss.Center,
			errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
	}

	hint := helpStyle.Render("Press ? for help")
	hintCentered := lipgloss.PlaceHorizontal(m.width, lipgloss.Center, hint)

	elements := []string{
		headerCentered,
		"",
		panelsCentered,
		nowPlayingCentered,
	}
	if errorCentered != "" {
		elements = append(elements, errorCentered)
	}
	elements = append(elements, hintCentered)

	view := lipgloss.JoinVertical(lipgloss.Center, elements...)

	if m.showHelp {
		helpContent := strings.Join([]string{
			"─────────── Keybinds ───────────",
			"",
			"[Space]    Play / Pause",
			"[Enter]    Select item",
			"[↑/↓]      Navigate list",
			"[Tab]      Switch panel",
			"[H/←]      Previous album",
			"[L/→]      Next album",
			"[N]        Next track",
			"[P]        Previous track",
			"[Q]        Quit",
			"[?]        Toggle help",
			"[Esc]      Close help",
			"",
			"────────────────────────────────",
		}, "\n")

		modalStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(1, 3).
			Background(colorBackground)

		modal := modalStyle.Render(helpContent)
		view = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modal)
	}

	return view
}

func (m Model) renderNowPlaying() string {
	if m.isLoading {
		content := lipgloss.NewStyle().
			Foreground(colorSecondary).
			Bold(true).
			Align(lipgloss.Center).
			Render("⟳ Loading track...")
		return nowPlayingStyle.Width(m.width - 10).Align(lipgloss.Center).Render(content)
	}

	if m.currentTrack == nil {
		content := lipgloss.NewStyle().
			Foreground(colorSubtext).
			Align(lipgloss.Center).
			Render("♪ No track playing ♪")
		return nowPlayingStyle.Width(m.width - 10).Align(lipgloss.Center).Render(content)
	}

	icon := "▶"
	if !m.isPlaying {
		icon = "⏸"
	}
	iconStyle := lipgloss.NewStyle().
		Foreground(colorText).
		Bold(true).
		Padding(0, 1)
	largeIcon := iconStyle.Render(icon)

	trackStyle := lipgloss.NewStyle().Bold(true).Foreground(colorText)
	artistStyle := lipgloss.NewStyle().Foreground(colorSubtext)
	trackInfo := trackStyle.Render(m.currentTrack.Name) + "  " + artistStyle.Render(m.currentTrack.Artist)
	posStr := formatDuration(m.position)
	durStr := formatDuration(m.duration)
	timeStr := fmt.Sprintf("%s / %s", posStr, durStr)

	content := fmt.Sprintf("%s  %s\n%s  %s", largeIcon, trackInfo, m.progressBar.View(), timeStr)

	return nowPlayingStyle.Width(m.width - 10).Align(lipgloss.Center).Render(content)
}

func formatDuration(d time.Duration) string {
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%d:%02d", m, s)
}

type musicItem struct {
	jellyfin.MusicItem
}

func (i musicItem) FilterValue() string { return i.Name }
func (i musicItem) Title() string       { return i.Name }
func (i musicItem) Description() string { return "" }

type musicDelegate struct{}

func (d musicDelegate) Height() int                             { return 1 }
func (d musicDelegate) Spacing() int                            { return 0 }
func (d musicDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d musicDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(musicItem)
	if !ok {
		return
	}

	if index == m.Index() {
		fmt.Fprint(w, selectedListItemStyle.Render("> "+i.Name))
	} else {
		fmt.Fprint(w, listItemStyle.Render("  "+i.Name))
	}
}

type albumHeader struct {
	name string
}

func (a albumHeader) FilterValue() string { return "" }
func (a albumHeader) Title() string       { return a.name }
func (a albumHeader) Description() string { return "" }

type trackItem struct {
	jellyfin.MusicItem
	queueIndex int
}

func (t trackItem) FilterValue() string { return t.Name }
func (t trackItem) Title() string       { return t.Name }
func (t trackItem) Description() string { return "" }

type trackDelegate struct{}

func (d trackDelegate) Height() int                             { return 1 }
func (d trackDelegate) Spacing() int                            { return 0 }
func (d trackDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d trackDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	switch item := listItem.(type) {
	case albumHeader:
		fmt.Fprint(w, albumHeaderStyle.Render(item.name))
	case trackItem:
		if index == m.Index() {
			fmt.Fprint(w, selectedListItemStyle.Render("> "+item.Name))
		} else {
			fmt.Fprint(w, listItemStyle.Render("  - "+item.Name))
		}
	}
}
