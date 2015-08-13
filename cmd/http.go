package envy

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/progrium/envy/pkg/hterm"
)

func init() {
	http.HandleFunc("/u/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) < 3 {
			http.NotFound(w, r)
			return
		}
		pathUser := parts[2]
		var pathEnv, sshUser string
		if len(parts) > 3 && parts[3] != "hterm" {
			pathEnv = parts[3]
			sshUser = pathUser + "+" + pathEnv
		} else {
			sshUser = pathUser
		}
		// passthrough auth for hterm. use cookie to do this right
		if !strings.Contains(r.URL.Path, "hterm") {
			user, passwd, ok := r.BasicAuth()
			if !ok || user != pathUser || !githubUserAuth(user, passwd) {
				w.Header().Set("WWW-Authenticate", fmt.Sprintf("Basic realm=\"%s\"", pathUser))
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				log.Println("auth[http]: not allowing", user)
				return
			}
			log.Println("auth[http]: allowing", user)
		}
		w.Header().Set("Hterm-Title", "Envy Term")
		hterm.Handle(w, r, func(args string) *hterm.Pty {
			cmd := exec.Command("/bin/enter", parts[2])
			cmd.Env = os.Environ()
			cmd.Env = append(cmd.Env, fmt.Sprintf("USER=%s", sshUser))
			pty, err := hterm.NewPty(cmd)
			if err != nil {
				log.Fatal(err)
			}
			return pty
		})
	})
}
