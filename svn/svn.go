package svn

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"log/slog"
	"strings"

	"os/exec"
)

type Service interface {
	Init()
	CurrentInfo() RepoInfo
	FetchInfo() error
	CurrentStatus() *RepoStatus
	FetchStatus() error
	StagePath(string) error
	UnstagePath(string) error
	FetchDiff(string) error
	GetDiff(string) []string
	GetPathStatus(SectionIdx, int) (PathStatus, error)
	ToggleSectionExpand(SectionIdx) error
	TogglePathExpand(SectionIdx, int) error
}

type RealService struct {
	RepoInfo   RepoInfo
	RepoStatus RepoStatus
	Logger     *slog.Logger
	diffCache  map[string][]string
}

func (svc *RealService) Init() {
	svc.Logger.Info("RealService.Init()")

	for i := SectionIdx(0); i < NumSections; i++ {
		svc.RepoStatus.Sections[i].Title = SectionTitles[i]
		svc.RepoStatus.Sections[i].Expanded = true
	}
	svc.RepoStatus.Sections[SectionUnversioned].Expanded = false
	if svc.diffCache == nil {
		svc.diffCache = make(map[string][]string)
	}
}

func (svc *RealService) CurrentInfo() RepoInfo {
	return svc.RepoInfo
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

func (svc *RealService) GetPathStatus(si SectionIdx, idx int) (PathStatus, error) {
	if si < 0 || si >= NumSections {
		return PathStatus{}, fmt.Errorf("GetPath with out of bounds section id called")
	}
	if idx < 0 || idx >= len(svc.RepoStatus.Sections[si].Paths) {
		return PathStatus{}, fmt.Errorf("GetPath with out of bounds idx called")
	}
	return svc.RepoStatus.Sections[si].Paths[idx], nil
}

func (svc *RealService) StagePath(path string) error {
	svc.Logger.Info("StagePath called", "path", path)
	if path == "" {
		return nil
	}
	cmd := exec.Command(
		"svn", "--non-interactive",
		"changelist", "staged", path)

	_, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Error running svn changelist staged: %w", err)
	}

	return nil
}

func (svc *RealService) UnstagePath(path string) error {
	svc.Logger.Info("UnstagePath called", "path", path)
	if path == "" {
		return nil
	}
	cmd := exec.Command(
		"svn", "--non-interactive",
		"changelist", "--remove", path)

	_, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Error running svn changelist --remove %s: %w", path, err)
	}

	return nil
}

func (svc *RealService) FetchDiff(path string) error {
	svc.Logger.Info("FetchDiff called", "path", path)
	if path == "" {
		return fmt.Errorf("Empty path provided to diff")
	}

	if _, ok := svc.diffCache[path]; ok {
		return nil
	}

	svc.Logger.Info("diff not in diffCache, fetching with svn diff command")
	cmd := exec.Command(
		"svn", "--non-interactive",
		"diff", path)

	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Error running svn diff %s: %w", path, err)
	}
	out = bytes.ReplaceAll(out, []byte("\r\n"), []byte("\n")) // normalize line endings

	sep := []byte("====\n")
	idx := bytes.Index(out, sep)
	if idx < 0 {
		return fmt.Errorf("separator for diff not found")
	}

	changeSep := []byte("\n@@")
	idx = bytes.Index(out, changeSep)
	if idx < 0 {
		empty := []string{""} // empty string so we can still toggle expand
		svc.diffCache[path] = empty
		return nil
	}

	diff := string(out[idx+1:])
	lines := strings.Split(diff, "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	svc.diffCache[path] = lines

	return nil
}

func (svc *RealService) GetDiff(path string) []string {
	return svc.diffCache[path]
}

func (svc *RealService) ToggleSectionExpand(si SectionIdx) error {
	svc.Logger.Info("ToggleSectionExpand called", "section", si)
	if si < 0 || si >= NumSections {
		return fmt.Errorf("ToggleSectionExpand called with out of bounds section id")
	}

	svc.RepoStatus.Sections[si].Expanded = !svc.RepoStatus.Sections[si].Expanded
	return nil
}

func (svc *RealService) TogglePathExpand(si SectionIdx, pathIdx int) error {
	svc.Logger.Info("TogglePathExpand called", "section", si, "pathIdx", pathIdx)
	if si < 0 || si >= NumSections {
		return fmt.Errorf("GetPath with out of bounds section id called")
	}
	if pathIdx < 0 || pathIdx >= len(svc.RepoStatus.Sections[si].Paths) {
		return fmt.Errorf("GetPath with out of bounds idx called")
	}
	svc.RepoStatus.Sections[si].Paths[pathIdx].Expanded =
		!svc.RepoStatus.Sections[si].Paths[pathIdx].Expanded
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
	Title    string
	Paths    []PathStatus
	Expanded bool
}

type PathStatus struct {
	Path     string
	Status   rune
	Expanded bool
}

type RepoStatus struct {
	Sections [NumSections]Section
}

func NewRepoStatus() RepoStatus {
	return RepoStatus{
		Sections: [NumSections]Section{
			{Title: "Unversioned"},
			{Title: "Unstaged"},
			{Title: "Staged"},
			{Title: "Ignored"},
			{Title: "Issues"},
		},
	}
}

func (rs *RepoStatus) Len(si SectionIdx) int {
	if si < 0 || si >= NumSections {
		return 0
	}
	return len(rs.Sections[si].Paths)
}

func (rs *RepoStatus) NextNonEmptySection(curr SectionIdx) (next SectionIdx, found bool) {
	for sec := curr + 1; sec < NumSections; sec++ {
		if len(rs.Sections[sec].Paths) > 0 {
			return sec, true
		}
	}
	return 0, false
}

func (rs *RepoStatus) PrevNonEmptySection(curr SectionIdx) (prev SectionIdx, found bool) {
	for sec := curr - 1; sec >= 0; sec-- {
		if len(rs.Sections[sec].Paths) > 0 {
			return sec, true
		}
	}
	return 0, false
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
	Revision int      `xml:"revision,attr"`
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
