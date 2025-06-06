package svn

import (
	"encoding/xml"
	"fmt"

	"os/exec"
)

type Service interface {
	CurrentInfo() RepoInfo
	FetchInfo() error
	FetchStatus() error
}

type RealService struct {
	workingPath string
	remoteURL   string
	revision    uint32
}

func (svc *RealService) CurrentInfo() RepoInfo {
	return RepoInfo{
		WorkingPath: svc.workingPath,
		RemoteURL:   svc.remoteURL,
		Revision:    svc.revision,
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

	svc.workingPath = infoXML.Entry.WCInfo.WCAbspath
	svc.remoteURL = infoXML.Entry.URL
	svc.revision = infoXML.Entry.Revision

	return nil
}

func (svc *RealService) FetchStatus() error {
	cmd := exec.Command(
		"svn", "--non-interactive",
		"status", "C:/Code/GitHub/textual-test/", "--xml")

	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Error running svn info: %w", err)
	}

	var statusXML StatusXML
	if err := xml.Unmarshal(out, &statusXML); err != nil {
		return fmt.Errorf("error unmarshalling svn info: %w", err)
	}

	// TODO: Just print now to test parsing is correct
	for _, entry := range statusXML.Target.Entries {
		println(entry.Path, " - ", entry.WCStatus.Status)
	}

	return nil
}

// SVN STATUS XML Structs
type RepoStatus struct {
}

type StatusXML struct {
	XMLName xml.Name  `xml:"status"`
	Target  TargetXML `xml:"target"`
}

type TargetXML struct {
	XMLName xml.Name         `xml:"target"`
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
