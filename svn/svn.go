package svn

import (
	"encoding/xml"
	"fmt"

	"os/exec"
)

type Service interface {
	CurrentInfo() RepoInfo
	FetchInfo() error
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
	cmd := exec.Command("svn", "info", "C:/Code/GitHub/textual-test/", "--xml")

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

type RepoInfo struct {
	WorkingPath string
	RemoteURL   string
	Revision    uint32
}

type InfoXML struct {
	XMLName xml.Name `xml:"info"`
	Entry   EntryXML `xml:"entry"`
}

type EntryXML struct {
	XMLName  xml.Name `xml:"entry"`
	URL      string   `xml:"url"`
	Revision uint32   `xml:"revision,attr"`
	WCInfo   WCInfo   `xml:"wc-info"`
}

type WCInfo struct {
	XMLName   xml.Name `xml:"wc-info"`
	WCAbspath string   `xml:"wcroot-abspath"`
}
