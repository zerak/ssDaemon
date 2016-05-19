// Copyright 2013 bee authors
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

// based on bee
package main

import (
	"fmt"
	"os"
	path "path/filepath"
	"strings"
)

var cmdNew = &Command{
	UsageLine: "new [servername]",
	Short:     "Create a serverFramework application",
	Long: `
Creates a server based on serverFramework for the given server name in the current directory.

The command 'new' creates a folder named [servername] and inside the folder deploy
the following files/directories structure:

    |- conf
        |-  app.conf
    |- msgs
         |- default.go
    |- fmt.sh
    |- main.go
    |- README.md
`,
}

func init() {
	cmdNew.Run = createApp
}

func createApp(cmd *Command, args []string) int {
	curpath, _ := os.Getwd()
	if len(args) != 1 {
		ColorLog("[ERRO] Argument [appname] is missing\n")
		os.Exit(2)
	}

	gopath := os.Getenv("GOPATH")
	Debugf("gopath:%s", gopath)
	if gopath == "" {
		ColorLog("[ERRO] $GOPATH not found\n")
		ColorLog("[HINT] Set $GOPATH in your environment vairables\n")
		os.Exit(2)
	}
	haspath := false
	appsrcpath := ""

	wgopath := path.SplitList(gopath)
	for _, wg := range wgopath {

		wg = path.Join(wg, "src")

		if strings.HasPrefix(strings.ToLower(curpath), strings.ToLower(wg)) {
			haspath = true
			appsrcpath = wg
			break
		}

		wg, _ = path.EvalSymlinks(wg)

		if strings.HasPrefix(strings.ToLower(curpath), strings.ToLower(wg)) {
			haspath = true
			appsrcpath = wg
			break
		}

	}

	if !haspath {
		ColorLog("[ERRO] Unable to create an application outside of $GOPATH%ssrc(%s%ssrc)\n", string(path.Separator), gopath, string(path.Separator))
		ColorLog("[HINT] Change your work directory by `cd ($GOPATH%ssrc)`\n", string(path.Separator))
		os.Exit(2)
	}

	apppath := path.Join(curpath, args[0])

	if isExist(apppath) {
		ColorLog("[ERRO] Path (%s) already exists\n", apppath)
		ColorLog("[WARN] Do you want to overwrite it? [yes|no]]")
		if !askForConfirmation() {
			os.Exit(2)
		}
	}

	fmt.Println("[INFO] Creating application...")

	os.MkdirAll(apppath, 0755)
	fmt.Println(apppath + string(path.Separator))
	os.Mkdir(path.Join(apppath, "conf"), 0755)
	fmt.Println(path.Join(apppath, "conf") + string(path.Separator))
	os.Mkdir(path.Join(apppath, "msgs"), 0755)
	fmt.Println(path.Join(apppath, "msgs") + string(path.Separator))
	os.Mkdir(path.Join(apppath, "models"), 0755)
	fmt.Println(path.Join(apppath, "models") + string(path.Separator))
	fmt.Println(path.Join(apppath, "conf", "app.conf"))
	writetofile(path.Join(apppath, "conf", "app.conf"), strings.Replace(appconf, "{{.Appname}}", args[0], -1))

	fmt.Println(path.Join(apppath, "models", "tables.go"))
	writetofile(path.Join(apppath, "models", "tables.go"), models)

	fmt.Println(path.Join(apppath, "msgs", "msglogin.go"))
	writetofile(path.Join(apppath, "msgs", "msglogin.go"), strings.Replace(msgs, "{{.Appname}}", strings.Join(strings.Split(apppath[len(appsrcpath)+1:], string(path.Separator)), "/"), -1))

	fmt.Println(path.Join(apppath, "main.go"))
	writetofile(path.Join(apppath, "main.go"), strings.Replace(maingo, "{{.Appname}}", strings.Join(strings.Split(apppath[len(appsrcpath)+1:], string(path.Separator)), "/"), -1))

	fmt.Println(path.Join(apppath,"fmt.sh"))
	writetofile(path.Join(apppath,"fmt.sh"),fmtsh)
	os.Chmod(path.Join(apppath,"fmt.sh"),os.FileMode(0775))

	fmt.Println(path.Join(apppath,"build.sh"))
	writetofile(path.Join(apppath,"build.sh"),strings.Replace(build, "{{.Appname}}", args[0], -1))
	os.Chmod(path.Join(apppath,"build.sh"),os.FileMode(0775))

	fmt.Println(path.Join(apppath,"README.md"))
	writetofile(path.Join(apppath,"README.md"),strings.Replace(readme, "{{.Appname}}", args[0], -1))

	ColorLog("[SUCC] New server successfully created!\n")
	return 0
}

var appconf = `# server basic conf
AppName = {{.Appname}}

# run mode dev|prod
RunMode = prod

RouterCaseSensitive = false
ServerName = server01
RecoverPanic = true
MaxMemory = 1 << 26
EnableErrorsShow = true

# server listen ip:port
TCPAddr = 127.0.0.1
TCPPort = 60060

# client buffer size
MsgSize = 10000

# server monitor conf
ServerTimeOut = 0
ListenTCP4 = false
EnableHTTP = false
HTTPAddr = 127.0.0.1
HTTPPort = 60100
EnableAdmin = true
AdminAddr = 127.0.0.1
AdminPort = 60200

# server log conf
LogAccessLogs = false
LogFileLineNum = true
LogOutputs = file, {"filename":"{{.Appname}}.log","level":7,"maxlines":0,"maxsize":0,"daily":true,"maxdays":1}

# DB conf
DBUser = user
DBPW = pw
DBAddr = localhost
DBPort = 3306
DBName = testDb
`

var models = `package models

import (
	"time"

	"github.com/astaxie/beego/orm"
)

type Table struct {
	ID   int       "auto"
	Time time.Time "type(datetime)"
}

type Mtable struct {
	ID   int       "auto"
	Time time.Time "type(datetime)"
}

func init() {
	orm.RegisterModel(new(Table))
	orm.RegisterModel(new(Mtable))
}

`

var maingo = `package main

import (
	"fmt"
	"runtime"

	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"

	"github.com/TaXingTianJi/serverFramework/core"
	"github.com/TaXingTianJi/serverFramework/utils"

	_ "{{.Appname}}/models"
	_ "{{.Appname}}/msgs"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	core.ServerApp.Version("{{.Appname}}")

	// orm
	core.SConfig.DBConf.User = "user"
	core.SConfig.DBConf.PW = "pw"
	//core.SConfig.DBConf.Addr = "localhost"
	//core.SConfig.DBConf.Port = 3306
	//core.SConfig.DBConf.DB = "testDb"
	str := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8",
		core.SConfig.DBConf.User, core.SConfig.DBConf.PW,
		core.SConfig.DBConf.Addr, core.SConfig.DBConf.Port,
		core.SConfig.DBConf.DB)
	orm.RegisterDriver("mysql", orm.DRMySQL)
	orm.RegisterDataBase("default", "mysql", str)
	orm.SetMaxIdleConns("default", 30)
	orm.SetMaxOpenConns("default", 30)

	core.Run()
	//core.Run("127.0.0.1:60060")
	//core.Run("localhost")
	//core.Run(":60060")

	var wg utils.WaitGroupWrapper
	wg.Wrap(func() {
		serverRoom()
	})
	wg.Wait()
}

func serverRoom() {
	for {

	}
}

`

var msgs = `package msgs

import (
	"fmt"
	"strconv"

	"github.com/astaxie/beego/orm"

	. "github.com/TaXingTianJi/serverFramework/client"
	. "github.com/TaXingTianJi/serverFramework/core"
	. "github.com/TaXingTianJi/serverFramework/protocol"

	. "{{.Appname}}/models"
)

type MsgLogin struct {
}

func init() {
	RegisterMsg(strconv.Itoa(10011), &MsgLogin{})
}

func (m *MsgLogin) ProcessMsg(p Protocol, client Client, msg *Message) {
	ServerLogger.Info("cid[%v] msg login", client.GetID())

	o := orm.NewOrm()
	o.Using("default")
	//u := new(Mtable)
	u := new(Table)
	u.Time = msg.Timestamp
	_, err := o.Insert(u)
	if err != nil {
		ServerLogger.Error("login error %v", err)
		_, err := p.Send(client, []byte("s2c login error"))
		if err != nil {
			err = fmt.Errorf("failed to send response ->%s", err)
			client.Exit()
		}
	} else {
		e := o.Read(u)
		if e != nil {
			ServerLogger.Info("read", u.ID, u.Time)
		}

		_, err := p.Send(client, []byte("s2c login succ"))
		if err != nil {
			err = fmt.Errorf("failed to send response ->%s", err)
			client.Exit()
		}
	}

}


`

var fmtsh = `#!/bin/bash
#find . -name "*.go" | xargs goimports -w
find . -name "*.go" | xargs gofmt -w
`

var readme = `###		{{.Appname}} based on serverFramework		###

* {{.Appname}}
`

var build = `go build -o {{.Appname}} *.go
`

func writetofile(filename, content string) {
	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	f.WriteString(content)
}
