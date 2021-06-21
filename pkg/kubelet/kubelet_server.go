package kubelet

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"gopkg.in/yaml.v2"
	"k8s-firstcommit/pkg/api"
)

type KubeletServer struct {
	Kubelet       *Kubelet
	UpdateChannel chan api.ContainerManifest
}

func (s *KubeletServer) error(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "Internal Error: %#v", err)
}

func (s *KubeletServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	u, err := url.ParseRequestURI(req.RequestURI)
	if err != nil {
		s.error(w, err)
		return
	}
	switch {
	case u.Path == "/container":
		defer req.Body.Close()
		data, err := ioutil.ReadAll(req.Body)
		if err != nil {
			s.error(w, err)
			return
		}
		var manifest api.ContainerManifest
		err = yaml.Unmarshal(data, &manifest)
		if err != nil {
			s.error(w, err)
			return
		}
		s.UpdateChannel <- manifest
	case u.Path == "/containerInfo":
		container := u.Query().Get("container")
		if len(container) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Missing container query arg.")
			return
		}
		id, err := s.Kubelet.GetContainerID(container)
		body, err := s.Kubelet.GetContainerInfo(id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Internal Error: %#v", err)
			return
		}
		w.Header().Add("Content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, body)
	default:
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Not found.")
	}
}
