package svn

import (
	"encoding/xml"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"os/exec"
)

type Service interface {
	CurrentInfo() RepoInfo
	FetchInfo() error
	CurrentStatus() *RepoStatus
	FetchStatus() error
}

type FetchInfoMsg struct{}
type FetchStatusMsg struct{}

type RefreshInfoMsg struct{}
type RefreshStatusMsg struct{}

type RealService struct {
	RepoInfo   RepoInfo
	RepoStatus RepoStatus
}

func (svc *RealService) CurrentInfo() RepoInfo {
	return svc.RepoInfo
}

func FetchInfoCmd(s Service) tea.Cmd {
	return func() tea.Msg {
		if err := s.FetchInfo(); err != nil {
			return nil
		}
		return RefreshInfoMsg{}
	}
}

func FetchStatusCmd(s Service) tea.Cmd {
	return func() tea.Msg {
		if err := s.FetchStatus(); err != nil {
			return nil
		}
		return RefreshStatusMsg{}
	}
}

func (svc *RealService) FetchInfo() error {
	cmd := exec.Command(
		"svn", "--non-interactive",
		"info", "C:/Code/GitHub/textual-test/", "--xml")

	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Error running svn info: %w", err)
	}

	var infoXML InfoXML
	if err := xml.Unmarshal(out, &infoXML); err != nil {
		return fmt.Errorf("error unmarshalling svn info: %w", err)
	}

	svc.RepoInfo.WorkingPath = infoXML.Entry.WCInfo.WCAbspath
	svc.RepoInfo.RemoteURL = infoXML.Entry.URL
	svc.RepoInfo.Revision = infoXML.Entry.Revision

	return nil
}

func (svc *RealService) CurrentStatus() *RepoStatus {
	return &svc.RepoStatus
}

func entryToPathStatus(entry StatusEntryXML) (PathStatus, error) {
	statusRune, res := StatusToRune(entry.WCStatus.Status)
	if !res {
		err := fmt.Errorf("Invalid status %s in path %s", entry.WCStatus.Status, entry.Path)
		return PathStatus{}, err
	}
	return PathStatus{Path: entry.Path, Status: statusRune}, nil
}

func (svc *RealService) FetchStatus() error {
	svc.RepoStatus.Clear()

	cmd := exec.Command(
		"svn", "--non-interactive",
		"status", "C:/Code/GitHub/textual-test/", "--xml")

	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("error running svn status: %w", err)
	}

	var statusXML StatusXML
	if err := xml.Unmarshal(out, &statusXML); err != nil {
		return fmt.Errorf("error unmarshalling svn status: %w", err)
	}

	for _, entry := range statusXML.Target.Entries {
		ps, err := entryToPathStatus(entry)
		if err != nil {
			return err
		}

		switch entry.WCStatus.Status {
		case "unversioned":
			svc.RepoStatus.Append(SectionUnversioned, ps)
		case "added", "deleted", "modified", "missing", "replaced":
			svc.RepoStatus.Append(SectionUnstaged, ps)
		case "conflicted", "external", "obstructed":
			svc.RepoStatus.Append(SectionIssues, ps)
		case "ignored":
			svc.RepoStatus.Append(SectionIgnored, ps)
		}
	}

	for _, cl := range statusXML.ChangeLists {
		if cl.Name == "staged" {
			for _, entry := range cl.Entries {
				ps, err := entryToPathStatus(entry)
				if err != nil {
					return err
				}
				svc.RepoStatus.Append(SectionStaged, ps)
			}
		}
	}

	return nil
}

func StatusToRune(status string) (rune, bool) {
	switch status {
	case "added":
		return 'A', true
	case "conflicted":
		return 'C', true
	case "deleted":
		return 'D', true
	case "ignored":
		return 'I', true
	case "modified":
		return 'M', true
	case "replaced":
		return 'R', true
	case "external":
		return 'X', true
	case "unversioned":
		return '?', true
	case "missing":
		return '!', true
	case "obstructed":
		return '~', true
	default:
		return ' ', false
	}
}

type SectionIdx int

const (
	SectionUnversioned SectionIdx = iota
	SectionUnstaged
	SectionStaged
	SectionIgnored
	SectionIssues

	NumSections
)

var SectionTitles = [NumSections]string{
	"Unversioned",
	"Unstaged",
	"Staged",
	"Ignored",
	"Issues",
}

type Section struct {
	Title     string
	Paths     []PathStatus
	Collapsed bool
}

type PathStatus struct {
	Path   string
	Status rune
}

type RepoStatus struct {
	Sections [NumSections]Section
}

func NewRepoStatus() RepoStatus {
	return RepoStatus{
		Sections: [NumSections]Section{
			Section{Title: "Unversioned"},
			Section{Title: "Unstaged"},
			Section{Title: "Staged"},
			Section{Title: "Ignored"},
			Section{Title: "Issues"},
		},
	}
}

func (rs *RepoStatus) Unversioned() Section {
	return rs.Sections[SectionUnversioned]
}

func (rs *RepoStatus) Unstaged() Section {
	return rs.Sections[SectionUnstaged]
}

func (rs *RepoStatus) Staged() Section {
	return rs.Sections[SectionStaged]
}

func (rs *RepoStatus) Ignored() Section {
	return rs.Sections[SectionIgnored]
}

func (rs *RepoStatus) Issues() Section {
	return rs.Sections[SectionIssues]
}

func (rs *RepoStatus) Append(sec SectionIdx, ps PathStatus) {
	rs.Sections[sec].Paths = append(rs.Sections[sec].Paths, ps)
}

func (rs *RepoStatus) Clear() {
	for i := range rs.Sections {
		rs.Sections[i].Paths = rs.Sections[i].Paths[:0]
	}
}

// SVN STATUS XML Structs

type StatusXML struct {
	XMLName     xml.Name        `xml:"status"`
	Target      TargetXML       `xml:"target"`
	ChangeLists []ChangeListXML `xml:"changelist"`
}

type TargetXML struct {
	XMLName xml.Name         `xml:"target"`
	Entries []StatusEntryXML `xml:"entry"`
}

type ChangeListXML struct {
	XMLName xml.Name         `xml:"changelist"`
	Name    string           `xml:"name,attr"`
	Entries []StatusEntryXML `xml:"entry"`
}

type StatusEntryXML struct {
	XMLName  xml.Name    `xml:"entry"`
	Path     string      `xml:"path,attr"`
	WCStatus WCStatusXML `xml:"wc-status"`
	Commit   CommitXML   `xml:"commit"`
}

type CommitXML struct {
	XMLName  xml.Name `xml:"commit"`
	Revision uint32   `xml:"revision,attr"`
	Author   string   `xml:"author"`
}

type WCStatusXML struct {
	XMLName  xml.Name `xml:"wc-status"`
	Status   string   `xml:"item,attr"`
	Revision uint32   `xml:"revision,attr"`
}

// SVN INFO XML Structs

type RepoInfo struct {
	WorkingPath string
	RemoteURL   string
	Revision    uint32
}

type InfoXML struct {
	XMLName xml.Name     `xml:"info"`
	Entry   InfoEntryXML `xml:"entry"`
}

type InfoEntryXML struct {
	XMLName  xml.Name `xml:"entry"`
	URL      string   `xml:"url"`
	Revision uint32   `xml:"revision,attr"`
	WCInfo   WCInfo   `xml:"wc-info"`
}

type WCInfo struct {
	XMLName   xml.Name `xml:"wc-info"`
	WCAbspath string   `xml:"wcroot-abspath"`
}
