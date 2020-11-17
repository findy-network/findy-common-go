package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/duo-labs/webauthn.io/session"
	"github.com/duo-labs/webauthn/protocol"
	"github.com/duo-labs/webauthn/webauthn"
	"github.com/findy-network/findy-grpc/enclave"
	"github.com/findy-network/findy-grpc/utils"
	"github.com/golang/glog"
	"github.com/gorilla/mux"
	. "github.com/lainio/err2"
)

var (
	loggingFlags string
	port         int
	webAuthn     *webauthn.WebAuthn
	sessionStore *session.Store

	startServerCmd = flag.NewFlagSet("server", flag.ExitOnError)
)

func init() {
	startServerCmd.StringVar(&loggingFlags, "logging", "-logtostderr=true -v=2", "logging startup arguments")
	startServerCmd.IntVar(&port, "port", 8080, "server port")
}

func main() {
	defer CatchTrace(func(err error) {
		glog.Warningln("")
	})
	//startServerCmd.Parse(os.Args)
	utils.ParseLoggingArgs(loggingFlags)
	Check(enclave.InitSealedBox("fido-enclave.bolt"))

	var err error
	webAuthn, err = webauthn.New(&webauthn.Config{
		RPDisplayName: "OP Lab Corp.",                           // Display Name for your site
		RPID:          "localhost",                              // Generally the domain name for your site
		RPOrigin:      fmt.Sprintf("http://localhost:%d", port), // The origin URL for WebAuthn requests
		// RPIcon: "https://duo.com/logo.png", // Optional icon URL for your site
	})
	Check(err)

	sessionStore, err = session.NewStore()
	Check(err)

	r := mux.NewRouter()

	r.HandleFunc("/register/begin/{username}", BeginRegistration).Methods("GET")
	r.HandleFunc("/register/finish/{username}", FinishRegistration).Methods("POST")
	r.HandleFunc("/login/begin/{username}", BeginLogin).Methods("GET")
	r.HandleFunc("/login/finish/{username}", FinishLogin).Methods("POST")

	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./")))

	serverAddress := fmt.Sprintf(":%d", port)
	glog.Infoln("starting server at", serverAddress)
	glog.Infoln(http.ListenAndServe(serverAddress, r))
}

func BeginRegistration(w http.ResponseWriter, r *http.Request) {
	defer Catch(func(err error) {
		glog.Warningln("begin registration error:", err)
	})

	// get username/friendly name
	vars := mux.Vars(r)
	username, ok := vars["username"]
	glog.V(1).Infoln("begin registration", username)
	if !ok {
		jsonResponse(w, fmt.Errorf("must supply a valid username i.e. foo@bar.com"), http.StatusBadRequest)
		return
	}

	var err error

	// get user
	user, exists, err := enclave.GetUser(username)
	Check(err)

	// user doesn't exist, create new user
	if !exists {
		displayName := strings.Split(username, "@")[0]
		glog.V(2).Infoln("adding new user:", displayName)
		user = enclave.NewUser(username, displayName)
		Check(enclave.PutUser(user))
	}

	registerOptions := func(credCreationOpts *protocol.PublicKeyCredentialCreationOptions) {
		credCreationOpts.CredentialExcludeList = user.CredentialExcludeList()
		glog.V(1).Infoln("credexcl:", len(credCreationOpts.CredentialExcludeList))
	}

	defer Handle(&err, func() {
		glog.Errorln("error:", err)
		jsonResponse(w, err.Error(), http.StatusInternalServerError)
	})

	glog.V(1).Infoln("begin registration to webAuthn")
	options, sessionData, err := webAuthn.BeginRegistration(
		user,
		registerOptions,
	)
	Check(err)
	glog.V(1).Infof("sessionData: %v", sessionData)

	// store session data as marshaled JSON
	glog.V(1).Infoln("store session data")
	Check(sessionStore.SaveWebauthnSession("registration", sessionData, r, w))

	jsonResponse(w, options, http.StatusOK)
	glog.V(1).Infoln("begin registration end", username)
}

func FinishRegistration(w http.ResponseWriter, r *http.Request) {
	defer Catch(func(err error) {
		glog.Warningln("BEGIN finish registration:", err)
	})

	var err error

	// get username
	vars := mux.Vars(r)
	username := vars["username"]
	glog.V(1).Infoln("finish registration", username)

	defer Handle(&err, func() {
		glog.Errorln("error:", err)
		jsonResponse(w, err.Error(), http.StatusInternalServerError)
	})

	glog.Infoln("getting existing user", username)
	user, err := enclave.GetExistingUser(username)
	Check(err)

	glog.V(1).Infoln("get session data for registration")
	sessionData, err := sessionStore.GetWebauthnSession("registration", r)
	Check(err)

	glog.V(1).Infoln("call web authn finish registration and getting credential")
	credential, err := webAuthn.FinishRegistration(user, sessionData, r)
	Check(err)

	// Add needed data to User
	user.AddCredential(*credential)
	Check(user.AllocateCloudAgent())
	// Persist that data
	Check(enclave.PutUser(user))

	jsonResponse(w, "Registration Success", http.StatusOK)
	glog.V(1).Infoln("END finish registration", username)
}

func BeginLogin(w http.ResponseWriter, r *http.Request) {
	defer Catch(func(err error) {
		glog.Warningln("begin login", err)
	})

	var err error

	// get username
	vars := mux.Vars(r)
	username := vars["username"]
	glog.V(1).Infoln("BEGIN begin login", username)

	defer Handle(&err, func() {
		jsonResponse(w, err.Error(), http.StatusInternalServerError)
	})

	// get user
	user, err := enclave.GetExistingUser(username)
	Check(err)

	// generate PublicKeyCredentialRequestOptions, session data
	options, sessionData, err := webAuthn.BeginLogin(user)
	Check(err)

	// store session data as marshaled JSON
	err = sessionStore.SaveWebauthnSession("authentication", sessionData, r, w)
	Check(err)

	jsonResponse(w, options, http.StatusOK)
	glog.Infoln("END begin login", username)
}

func FinishLogin(w http.ResponseWriter, r *http.Request) {
	defer Catch(func(err error) {
		glog.Warningln("finish login error:", err)
	})

	var err error

	// get username
	vars := mux.Vars(r)
	username := vars["username"]
	glog.Infoln("BEGIN finish login:", username)

	defer Handle(&err, func() {
		jsonResponse(w, err.Error(), http.StatusInternalServerError)
	})

	// get user
	user, err := enclave.GetExistingUser(username)
	Check(err)

	// load the session data
	sessionData, err := sessionStore.GetWebauthnSession("authentication", r)
	Check(err)

	// in an actual implementation, we should perform additional checks on
	// the returned 'credential', i.e. check 'credential.Authenticator.CloneWarning'
	// and then increment the credentials counter
	_, err = webAuthn.FinishLogin(user, sessionData, r)
	Check(err)

	// handle successful login
	jsonResponse(w, "Login Success", http.StatusOK)
	glog.V(1).Infoln("END finish login", username)
}

// from: https://github.com/duo-labs/webauthn.io/blob/3f03b482d21476f6b9fb82b2bf1458ff61a61d41/server/response.go#L15
func jsonResponse(w http.ResponseWriter, d interface{}, c int) {
	dj, err := json.Marshal(d)
	if err != nil {
		http.Error(w, "Error creating JSON response", http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(c)
	glog.V(1).Infof("reply json:\n%s", dj)
	fmt.Fprintf(w, "%s", dj)
}