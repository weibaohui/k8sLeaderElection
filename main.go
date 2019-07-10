package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/google/uuid"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"time"

	"k8s.io/klog"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
)

func main() {
	go shareTask()
	config := leaderElectionConfig{
		Name:       "leaderElectionName",
		Namespace:  "default",
		ElectionID: "xxxxxx",
		Client:     getClient(),
		OnStartedLeading: func(i chan struct{}) {
			fmt.Println("leader now")
			go leader()
		},
		OnStoppedLeading: func() {
			fmt.Println("die die die")
		},
	}

	wait.Forever(func() {
		setupLeaderElection(&config)
	}, time.Second*5)
}
func shareTask() {
	for {
		time.Sleep(time.Second * 2)
		fmt.Println("shareTask")
	}
}
func leader() {
	for {
		time.Sleep(time.Second)
		fmt.Println("leader-leader-leader-leader-leader")
	}
}

type leaderElectionConfig struct {
	Name       string
	Namespace  string
	ElectionID string

	Client clientset.Interface

	OnStartedLeading func(chan struct{})
	OnStoppedLeading func()
}

func getClient() clientset.Interface {
	var config *rest.Config
	var err error
	config, err = rest.InClusterConfig()
	if err != nil {
		var path string
		flag.StringVar(&path, "kubeconfig", "", "")
		flag.Parse()
		config, err = clientcmd.BuildConfigFromFlags("", path)
		if err != nil {
			klog.Fatal(err)
		}
	}

	Client := clientset.NewForConfigOrDie(config)
	return Client

}

func setupLeaderElection(config *leaderElectionConfig) {
	var elector *leaderelection.LeaderElector

	// start a new context
	ctx := context.Background()

	var cancelContext context.CancelFunc

	var newLeaderCtx = func(ctx context.Context) context.CancelFunc {
		// allow to cancel the context in case we stop being the leader
		leaderCtx, cancel := context.WithCancel(ctx)
		go elector.Run(leaderCtx)
		return cancel
	}

	var stopCh chan struct{}
	callbacks := leaderelection.LeaderCallbacks{
		OnStartedLeading: func(ctx context.Context) {
			klog.V(2).Infof("I am the new leader")
			stopCh = make(chan struct{})

			if config.OnStartedLeading != nil {
				config.OnStartedLeading(stopCh)
			}
		},
		OnStoppedLeading: func() {
			klog.V(2).Info("I am not leader anymore")
			close(stopCh)

			// cancel the context
			cancelContext()

			cancelContext = newLeaderCtx(ctx)

			if config.OnStoppedLeading != nil {
				config.OnStoppedLeading()
			}
		},
		OnNewLeader: func(identity string) {
			klog.Infof("new leader elected: %v", identity)
		},
	}

	broadcaster := record.NewBroadcaster()
	hostname, _ := os.Hostname()

	recorder := broadcaster.NewRecorder(scheme.Scheme, apiv1.EventSource{
		Component: config.Name,
		Host:      hostname,
	})

	lock := resourcelock.ConfigMapLock{
		ConfigMapMeta: metav1.ObjectMeta{Namespace: config.Namespace, Name: config.ElectionID},
		Client:        config.Client.CoreV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity:      uuid.New().String(),
			EventRecorder: recorder,
		},
	}

	ttl := 30 * time.Second
	var err error

	elector, err = leaderelection.NewLeaderElector(leaderelection.LeaderElectionConfig{
		Lock:          &lock,
		LeaseDuration: ttl,
		RenewDeadline: ttl / 2,
		RetryPeriod:   ttl / 4,

		Callbacks: callbacks,
	})
	if err != nil {
		klog.Fatalf("unexpected error starting leader election: %v", err)
	}

	cancelContext = newLeaderCtx(ctx)
}
