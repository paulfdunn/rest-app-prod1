// prod1 is an example of how to base multiple products off of the same init/config functionality,
// using rest-app (a framework for a GO (GOLANG) based ReST APIs).
// This can be used as the basis for a GO based app, needing JWT user authentication,
// with logging and key/value store (KVS).
// rest-app is hosted at https://github.com/paulfdunn/rest-app; please see the repo
// for more information
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"

	"github.com/paulfdunn/authJWT"
	"github.com/paulfdunn/logh"
	"github.com/paulfdunn/osh/runtimeh"
	"github.com/paulfdunn/rest-app/common"
	"github.com/paulfdunn/rest-app/common/config"
)

const (
	appName = "prod1"
)

var (
	filepathsToDeleteOnReset = []string{}
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			errOut := fmt.Sprintf("panic: %+v\n%+v", err, string(debug.Stack()))
			fmt.Println(errOut)
			logh.Map[appName].Println(logh.Error, errOut)
			logh.ShutdownAll()
		}
	}()

	an := appName
	dcnfg := config.Config{AppName: &an, LogName: &an}

	// default to the executable path.
	exe, err := os.Executable()
	if err != nil {
		log.Fatalf("fatal: %s fatal: Could not find executable path.", runtimeh.SourceInfo())
	}
	ap := filepath.Dir(exe)
	dcnfg.AppPath = &ap

	tfp := filepath.Join(*dcnfg.AppPath, "/key/jwt.rsa.public")
	dcnfg.JWTKeyFilepath = &tfp
	ri := time.Minute
	dcnfg.JWTAuthRemoveInterval = &ri
	ti := time.Minute * 15
	dcnfg.JWTAuthTimeoutInterval = &ti
	// default password validation: 8-32 characters, 1 lower case, 1 upper case, 1 special, 1 number.
	dcnfg.PasswordValidation = []string{`^[\S]{8,32}$`, `[a-z]`, `[A-Z]`, `[!#$%'()*+,-.\\/:;=?@\[\]^_{|}~]`, `[0-9]`}
	common.ConfigInit(dcnfg, filepathsToDeleteOnReset)
	var cnfg config.Config
	if cnfg, err = config.Get(); err != nil {
		log.Fatalf("fatal: %s getting running config, error:%v", runtimeh.SourceInfo(), err)
	}
	logh.Map[appName].Printf(logh.Info, "Config: %s", cnfg)

	ac := authJWT.Config{
		AppName:                *cnfg.AppName,
		AuditLogName:           *cnfg.AuditLogName,
		DataSourceName:         *cnfg.DataSourceName,
		CreateRequiresAuth:     true,
		JWTAuthRemoveInterval:  *cnfg.JWTAuthRemoveInterval,
		JWTAuthTimeoutInterval: *cnfg.JWTAuthTimeoutInterval,
		JWTKeyFilepath:         *cnfg.JWTKeyFilepath,
		LogName:                *cnfg.LogName,
		PasswordValidation:     cnfg.PasswordValidation,
		PathCreate:             "/auth/create",
		PathDelete:             "/auth/delete",
		PathInfo:               "/auth/info",
		PathLogin:              "/auth/login",
		PathLogout:             "/auth/logout",
		PathLogoutAll:          "/auth/logout-all",
		PathRefresh:            "/auth/refresh",
	}
	mux := http.NewServeMux()
	common.OtherInit(ac, mux)
	if *cnfg.NewDataSource {
		initDataSource()
	}

	// Registering with the trailing slash means the naked path is redirected to this path.
	path := "/"
	mux.HandleFunc(path, authJWT.HandlerFuncAuthJWTWrapper(handler))
	logh.Map[appName].Printf(logh.Info, "Registered handler: %s\n", path)

	// blocking call
	cfp := filepath.Join(*dcnfg.AppPath, "/key/rest-app.crt")
	kfp := filepath.Join(*dcnfg.AppPath, "/key/rest-app.key")
	common.ListenAndServeTLS(appName, mux, fmt.Sprintf(":%d", *cnfg.HTTPSPort),
		10*time.Second, 10*time.Second, cfp, kfp)
}

func initDataSource() {
	// Create default auth
	em := "admin"
	pw := "P@ss!234"
	cred := authJWT.Credential{Email: &em, Password: &pw}
	if err := cred.AuthCreate(); err != nil {
		log.Fatalf("fatal: %s creating default account, error: %v", runtimeh.SourceInfo(), err)
	}
	logh.Map[appName].Printf(logh.Info, "Created default auth: %s\n", em)
}

func handler(w http.ResponseWriter, r *http.Request) {
	logh.Map[appName].Printf(logh.Info, "rest-app handler %v\n", *r)
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
		logh.Map[appName].Printf(logh.Error, "hostname error: %v\n", err)
	}
	w.Write([]byte(fmt.Sprintf("hostname: %s, rest-app - from github.com/paulfdunn/rest-app", hostname)))
}
