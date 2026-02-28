package checker

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func makePod(name, namespace string, phase corev1.PodPhase, containerStatuses []corev1.ContainerStatus) corev1.Pod {
	return corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Status: corev1.PodStatus{
			Phase:             phase,
			ContainerStatuses: containerStatuses,
		},
	}
}

func TestIsPodFailing_PhaseFailed(t *testing.T) {
	pod := makePod("test-pod", "openshift-monitoring", corev1.PodFailed, nil)
	if !isPodFailing(pod) {
		t.Error("expected pod with phase=Failed to be failing")
	}
}

func TestIsPodFailing_PhaseRunning_Healthy(t *testing.T) {
	pod := makePod("test-pod", "openshift-monitoring", corev1.PodRunning, []corev1.ContainerStatus{
		{
			Name:  "app",
			Ready: true,
			State: corev1.ContainerState{
				Running: &corev1.ContainerStateRunning{},
			},
		},
	})
	if isPodFailing(pod) {
		t.Error("expected running healthy pod to NOT be failing")
	}
}

func TestIsPodFailing_CrashLoopBackOff(t *testing.T) {
	pod := makePod("crash-pod", "openshift-monitoring", corev1.PodRunning, []corev1.ContainerStatus{
		{
			Name:  "app",
			Ready: false,
			State: corev1.ContainerState{
				Waiting: &corev1.ContainerStateWaiting{
					Reason: "CrashLoopBackOff",
				},
			},
		},
	})
	if !isPodFailing(pod) {
		t.Error("expected pod with CrashLoopBackOff container to be failing")
	}
}

func TestIsPodFailing_OOMKilled(t *testing.T) {
	pod := makePod("oom-pod", "openshift-monitoring", corev1.PodRunning, []corev1.ContainerStatus{
		{
			Name:  "app",
			Ready: false,
			State: corev1.ContainerState{
				Terminated: &corev1.ContainerStateTerminated{
					Reason: "OOMKilled",
				},
			},
		},
	})
	if !isPodFailing(pod) {
		t.Error("expected pod with OOMKilled container to be failing")
	}
}

func TestIsPodFailing_ErrorTerminated(t *testing.T) {
	pod := makePod("error-pod", "openshift-monitoring", corev1.PodRunning, []corev1.ContainerStatus{
		{
			Name:  "app",
			Ready: false,
			State: corev1.ContainerState{
				Terminated: &corev1.ContainerStateTerminated{
					Reason: "Error",
				},
			},
		},
	})
	if !isPodFailing(pod) {
		t.Error("expected pod with Error terminated container to be failing")
	}
}

func TestIsPodFailing_PhaseSucceeded(t *testing.T) {
	pod := makePod("completed-pod", "openshift-monitoring", corev1.PodSucceeded, nil)
	if isPodFailing(pod) {
		t.Error("expected pod with phase=Succeeded to NOT be failing")
	}
}

func TestIsPodFailing_PhasePending(t *testing.T) {
	pod := makePod("pending-pod", "openshift-monitoring", corev1.PodPending, nil)
	if isPodFailing(pod) {
		t.Error("expected pod with phase=Pending to NOT be failing")
	}
}

func TestIsContainerFailing_WaitingCrashLoopBackOff(t *testing.T) {
	cs := corev1.ContainerStatus{
		State: corev1.ContainerState{
			Waiting: &corev1.ContainerStateWaiting{Reason: "CrashLoopBackOff"},
		},
	}
	if !isContainerFailing(cs) {
		t.Error("expected CrashLoopBackOff waiting container to be failing")
	}
}

func TestIsContainerFailing_WaitingOtherReason(t *testing.T) {
	cs := corev1.ContainerStatus{
		State: corev1.ContainerState{
			Waiting: &corev1.ContainerStateWaiting{Reason: "ContainerCreating"},
		},
	}
	if isContainerFailing(cs) {
		t.Error("expected ContainerCreating waiting container to NOT be failing")
	}
}

func TestIsContainerFailing_TerminatedOOMKilled(t *testing.T) {
	cs := corev1.ContainerStatus{
		State: corev1.ContainerState{
			Terminated: &corev1.ContainerStateTerminated{Reason: "OOMKilled"},
		},
	}
	if !isContainerFailing(cs) {
		t.Error("expected OOMKilled terminated container to be failing")
	}
}

func TestIsContainerFailing_TerminatedCompleted(t *testing.T) {
	cs := corev1.ContainerStatus{
		State: corev1.ContainerState{
			Terminated: &corev1.ContainerStateTerminated{Reason: "Completed"},
		},
	}
	if isContainerFailing(cs) {
		t.Error("expected Completed terminated container to NOT be failing")
	}
}
