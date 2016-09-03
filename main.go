package main

import (
	"encoding/base64"
	"fmt"
	log "github.com/cihub/seelog"
	"github.com/satori/go.uuid"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"runtime/debug"
	"strings"
	"time"
)

// remove docker warnings eg "WARNING: Your kernel does not support swap limit capabilities, memory limited without swap."
func removeWarnings(s string) string {
	split := strings.Split(strings.Trim(string(s), "\n"), "\n")
	ret := []string{}
	for index, line := range split {
		if !strings.HasPrefix(line, "WARNING: ") {
			if index > 0 {
				log.Infof("Cutting junk from output: %v", strings.Join(split[0:index], "\n"))
			}
			ret = split[index:]
			break
		}
	}
	return strings.Join(ret, "\n")
}

func wrapper(f func(code []byte) ([]byte, error)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				debug.PrintStack()
				log.Errorf("Panic happened: %v", r)
				fmt.Fprintf(w, `{"error": "server panicked", "context": "%v"}`, r)
			}
		}()
		w.Header().Set("Access-Control-Allow-Origin", "*")
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"error":"failed to read request body", "context":"%v"}`, err)
			return
		}
		decoded := body // , err := base64.StdEncoding.DecodeString(string(body))
		//if err != nil {
		//	log.Debugf("Decode error: %v", err)
		//	fmt.Fprintf(w, `{"error":"failed to decode request", "context":"%v"}`, err)
		//	return
		//}
		output, err := f(decoded)
		if err != nil {
			log.Errorf("Error running code: output: %v, err: %v", string(output), err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"error":"failed to run code", "context":"%v", "output": "%v"}`, err, base64.StdEncoding.EncodeToString([]byte(removeWarnings(string(output)))))
			return
		}
		fmt.Fprintf(w, `{"output":"%v"}`, base64.StdEncoding.EncodeToString([]byte(removeWarnings(string(output)))))
	}
}

func node(code []byte) ([]byte, error) {
	return runCode("noderunner", code)
}

func haskell(code []byte) ([]byte, error) {
	return runCode("haskellrunner", code)
}

func shell(code []byte) ([]byte, error) {
	return runCode("shellrunner", code)
}

func imageToMemoryLimit(image string) string {
	mb := 1024 * 1024
	switch image {
	case "noderunner":
		return fmt.Sprintf("%v", 8*mb)
	case "haskellrunner":
		return fmt.Sprintf("%v", 20*mb)
	case "shellrunner":
		return fmt.Sprintf("%v", 15*mb)
	}
	return fmt.Sprintf("%v", 8*mb)
}

func imageToCpuLimit(image string) string {
	switch image {
	case "noderunner":
		return "12"
	}
	return "12"
}

func runCode(image string, code []byte) ([]byte, error) {
	id := uuid.NewV4().String()
	log.Infof("Running code \n<===\n%v\n===>\nin image %v in container with name %v", string(code), image, id)
	defer func() {
		outp, err := exec.Command("docker", "rm", id).CombinedOutput()
		log.Infof("Cleaning up container with name %v", id)
		if err != nil {
			log.Errorf("Error removing container, output: %v, err: %v", string(outp), err)
		}
	}()
	// kill if still alive
	go func(id string) {
		time.Sleep(30 * time.Second)
		outp, err := exec.Command("docker", "rm", "-f", id).CombinedOutput()
		log.Debug(string(outp), err)
	}(id)
	args := []string{"run", "--cpu-shares", imageToCpuLimit(image), "-m", imageToMemoryLimit(image), "--name", id, image, string(code)}
	log.Debugf("Executing: %v", strings.Join(args, " "))
	return exec.Command("docker", args...).CombinedOutput()
}

func main() {
	http.HandleFunc("/node", wrapper(node))
	http.HandleFunc("/haskell", wrapper(haskell))
	http.HandleFunc("/shell", wrapper(shell))
	log.Critical(http.ListenAndServe(":8080", nil))
}

// below it's all log related

const (
	defaultLogLevel  = "debug"
	systemConfigFile = "/etc/seelog.xml"
)

var logFormat = `
<seelog minlevel="%s">
	<outputs formatid="main">
		<console/>
	</outputs>
	<formats>
		<format id="main" format="%%UTCDate(2006-01-02 15:04:05.9999) %%LEVEL [%%File:%%Line %%FuncShort] %%Msg%%n"/>
	</formats>
</seelog>`

// Check if the level is valid
func isValid(level string) bool {
	if level == "trace" ||
		level == "debug" ||
		level == "info" ||
		level == "warn" ||
		level == "error" ||
		level == "critical" {
		return true
	}
	return false
}

func init() {
	var (
		logger log.LoggerInterface
		err    error
	)

	log_level := os.Getenv("LOG_LVL")
	if !isValid(log_level) {
		log_level = defaultLogLevel
	}
	// Check if there is a system-wide configuration file for the seelog
	// package; use it if so.
	_, err = os.Stat(systemConfigFile)
	if err == nil {
		logger, err = log.LoggerFromConfigAsFile(systemConfigFile)
	} else {
		// We could modify this as we like
		// But S6 seems to do the piping to a file
		cfg := fmt.Sprintf(logFormat, log_level)
		logger, err = log.LoggerFromConfigAsString(cfg)
	}
	if err != nil {
		panic(err)
	}
	log.ReplaceLogger(logger)
	log.Infof("Logger initialized using level: %s", log_level)
}
