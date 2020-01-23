// Copyright (c) 2019 Sorint.lab S.p.A.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"bytes"
	"crypto/tls"
	b64 "encoding/base64"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"github.com/amreo/ercole-agent-absence-checker/config"
	"github.com/amreo/ercole-agent-absence-checker/marshal"
	"github.com/ercole-io/ercole-agent/scheduler"
	"github.com/ercole-io/ercole-agent/scheduler/storage"

	"github.com/kardianos/service"
)

var logger service.Logger
var version = "latest"

type program struct{}

func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}
func (p *program) run() {
	configuration := config.ReadConfig()
	buildData(configuration) // first run

	memStorage := storage.NewMemoryStorage()
	scheduler := scheduler.New(memStorage)

	_, err := scheduler.RunEvery(time.Duration(configuration.Frequency)*time.Hour, buildData, configuration)

	if err != nil {
		log.Fatal("Error sending data", err)
	}

	scheduler.Start()
	scheduler.Wait()
}

func (p *program) Stop(s service.Service) error {
	return nil
}

func main() {

	svcConfig := &service.Config{
		Name:        "ErcoleAgent",
		DisplayName: "The Ercole Agent",
		Description: "Asset management agent from the Ercole project.",
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}
	logger, err = s.Logger(nil)
	if err != nil {
		log.Fatal(err)
	}
	err = s.Run()
	if err != nil {
		logger.Error(err)
	}

}

func buildData(configuration config.Configuration) {
	out := fetcher(configuration, "host-list")
	list := marshal.HostList(out)
	for _, l := range list {
		sendData(l, configuration)
	}
}

func sendData(hostname string, configuration config.Configuration) {

	log.Println("Checking host " + hostname + "...")

	client := &http.Client{}

	//Disable certificate validation if enableServerValidation is false
	if configuration.EnableServerValidation == false {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	req, err := http.NewRequest("POST", configuration.Serverurl+"/"+hostname, bytes.NewReader([]byte{}))
	auth := configuration.Serverusr + ":" + configuration.Serverpsw
	authEnc := b64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Add("Authorization", "Basic "+authEnc)
	resp, err := client.Do(req)

	sendResult := "FAILED"

	if err != nil {
		log.Println("Error sending data", err)
	} else {
		log.Println("Response status:", resp.Status)
		if resp.StatusCode == 200 {
			sendResult = "SUCCESS"
		}
		defer resp.Body.Close()
	}

	log.Println("Sending result:", sendResult)

}

func fetcher(configuration config.Configuration, fetcherName string, params ...string) []byte {
	var (
		cmd    *exec.Cmd
		err    error
		psexe  string
		stdout bytes.Buffer
		stderr bytes.Buffer
	)

	log.Println("Fetching", fetcherName)

	baseDir := config.GetBaseDir()

	if runtime.GOOS == "windows" {
		psexe, err = exec.LookPath("powershell.exe")
		if err != nil {
			log.Fatal(psexe)
		}
		if configuration.ForcePwshVersion == "0" {
			params = append([]string{"-ExecutionPolicy", "Bypass", "-File", baseDir + "\\fetch\\" + fetcherName}, params...)
		} else {
			params = append([]string{"-version", configuration.ForcePwshVersion, "-ExecutionPolicy", "Bypass", "-File", baseDir + "\\fetch\\win.ps1", "-s", fetcherName}, params...)
		}
		cmd = exec.Command(psexe, params...)
	} else {
		cmd = exec.Command(baseDir+"/fetch/"+fetcherName, params...)
	}

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if len(stderr.Bytes()) > 0 {
		log.Print(string(stderr.Bytes()))
	}

	return stdout.Bytes()
}
