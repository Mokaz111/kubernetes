/*
Copyright 2014 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package scheduler

//import (
//	"context"
//	"fmt"
//	"math/rand"
//	"sort"
//	"strings"
//
//	"time"
//
//	"k8s.io/klog/v2"
//
//	v1 "k8s.io/api/core/v1"
//
//	"k8s.io/kubernetes/pkg/scheduler/framework"
//	internalcache "k8s.io/kubernetes/pkg/scheduler/internal/cache"
//	utiltrace "k8s.io/utils/trace"
//)
//
//// FitError describes a fit error of a pod.
//type FitError struct {
//	Pod                   *v1.Pod
//	NumAllNodes           int
//	FilteredNodesStatuses framework.NodeToStatusMap
//}
//
//// ErrNoNodesAvailable is used to describe the error that no nodes available to schedule pods.
//
//const (
//	// NoNodeAvailableMsg is used to format message when no nodes available.
//	NoNodeAvailableMsg = "0/%v nodes are available"
//)
//
//// Error returns detailed information of why the pod failed to fit on each node
//func (f *FitError) Error() string {
//	reasons := make(map[string]int)
//	for _, status := range f.FilteredNodesStatuses {
//		for _, reason := range status.Reasons() {
//			reasons[reason]++
//		}
//	}
//
//	sortReasonsHistogram := func() []string {
//		var reasonStrings []string
//		for k, v := range reasons {
//			reasonStrings = append(reasonStrings, fmt.Sprintf("%v %v", v, k))
//		}
//		sort.Strings(reasonStrings)
//		return reasonStrings
//	}
//	reasonMsg := fmt.Sprintf(NoNodeAvailableMsg+": %v.", f.NumAllNodes, strings.Join(sortReasonsHistogram(), ", "))
//	return reasonMsg
//}
//
//// ScheduleAlgorithm is an interface implemented by things that know how to schedule pods
//// onto machines.
//// TODO: Rename this type.
//type ScheduleAlgorithm interface {
//	Filters
//	Schedule(context.Context, framework.Framework, *framework.CycleState, *v1.Pod) (scheduleResult ScheduleResult, err error)
//}
//
//// Filters is an interface to reuse the filter result
//type Filters interface {
//	Filter(context.Context, framework.Framework, *framework.CycleState, *v1.Pod) (scheduleResult ScheduleResult, err error)
//}
//
//// ScheduleResult represents the result of one pod scheduled. It will contain
//// the final selected Node, along with the selected intermediate information.
//type ScheduleResult struct {
//	// Number of nodes scheduler evaluated on one pod scheduled
//	EvaluatedNodes int
//	// Number of feasible nodes on one pod scheduled
//	FeasibleNodes int
//}
//
//type genericScheduler struct {
//	cache                    internalcache.Cache
//	nodeInfoSnapshot         *internalcache.Snapshot
//	percentageOfNodesToScore int32
//	nextStartNodeIndex       int
//}
//
//// snapshot snapshots scheduler cache and node infos for all fit and priority
//// functions.
//func (g *genericScheduler) snapshot() error {
//	// Used for all fit and priority funcs.
//	return g.cache.UpdateSnapshot(klog.Logger{}, g.nodeInfoSnapshot)
//}
//
//// Filter tries to schedule the given pod to one of the nodes in the node list.
//// If it succeeds, it will return the name of the node.
//// If it fails, it will return a FitError error with reasons.
//func (g *genericScheduler) Filter(ctx context.Context, fwk framework.Framework, state *framework.CycleState,
//	pod *v1.Pod) (result ScheduleResult, err error) {
//	trace := utiltrace.New("Scheduling", utiltrace.Field{Key: "namespace", Value: pod.Namespace}, utiltrace.Field{Key: "name", Value: pod.Name})
//	defer trace.LogIfLong(100 * time.Millisecond)
//
//	if err := g.snapshot(); err != nil {
//		return result, err
//	}
//	trace.Step("Snapshotting scheduler cache and node infos done")
//
//	if g.nodeInfoSnapshot.NumNodes() == 0 {
//		return result, ErrNoNodesAvailable
//	}
//
//	feasibleNodes, filteredNodesStatuses, err := g.findNodesThatFitPod(ctx, fwk, state, pod)
//	if err != nil {
//		return result, err
//	}
//	trace.Step("Computing predicates done")
//
//	if len(feasibleNodes) == 0 {
//		return result, &FitError{
//			Pod:                   pod,
//			NumAllNodes:           g.nodeInfoSnapshot.NumNodes(),
//			FilteredNodesStatuses: filteredNodesStatuses,
//		}
//	}
//
//	return ScheduleResult{
//		SuggestedHost:  feasibleNodes[0].Name,
//		EvaluatedNodes: 1 + len(filteredNodesStatuses),
//		FeasibleNodes:  1,
//	}, nil
//}
//
//// Schedule tries to schedule the given pod to one of the nodes in the node list.
//// If it succeeds, it will return the name of the node.
//// If it fails, it will return a FitError error with reasons.
//func (g *genericScheduler) Schedule(ctx context.Context, fwk framework.Framework, state *framework.CycleState, pod *v1.Pod) (result ScheduleResult, err error) {
//	trace := utiltrace.New("Scheduling", utiltrace.Field{Key: "namespace", Value: pod.Namespace}, utiltrace.Field{Key: "name", Value: pod.Name})
//	defer trace.LogIfLong(100 * time.Millisecond)
//
//	if err := g.snapshot(); err != nil {
//		return result, err
//	}
//	trace.Step("Snapshotting scheduler cache and node infos done")
//
//	if g.nodeInfoSnapshot.NumNodes() == 0 {
//		return result, ErrNoNodesAvailable
//	}
//
//	feasibleNodes, filteredNodesStatuses, err := g.findNodesThatFitPod(ctx, fwk, state, pod)
//	if err != nil {
//		return result, err
//	}
//	trace.Step("Computing predicates done")
//
//	if len(feasibleNodes) == 0 {
//		return result, &FitError{
//			Pod:                   pod,
//			NumAllNodes:           g.nodeInfoSnapshot.NumNodes(),
//			FilteredNodesStatuses: filteredNodesStatuses,
//		}
//	}
//
//	// When only one node after predicate, just use it.
//	if len(feasibleNodes) == 1 {
//		return ScheduleResult{
//			SuggestedHost:  feasibleNodes[0].Name,
//			EvaluatedNodes: 1 + len(filteredNodesStatuses),
//			FeasibleNodes:  1,
//		}, nil
//	}
//	trace.Step("Prioritizing done")
//
//	return ScheduleResult{
//		EvaluatedNodes: len(feasibleNodes) + len(filteredNodesStatuses),
//		FeasibleNodes:  len(feasibleNodes),
//	}, err
//}
//
//// selectHost takes a prioritized list of nodes and then picks one
//// in a reservoir sampling manner from the nodes that had the highest score.
//func (g *genericScheduler) selectHost(nodeScoreList framework.NodeScoreList) (string, error) {
//	if len(nodeScoreList) == 0 {
//		return "", fmt.Errorf("empty priorityList")
//	}
//	maxScore := nodeScoreList[0].Score
//	selected := nodeScoreList[0].Name
//	cntOfMaxScore := 1
//	for _, ns := range nodeScoreList[1:] {
//		if ns.Score > maxScore {
//			maxScore = ns.Score
//			selected = ns.Name
//			cntOfMaxScore = 1
//		} else if ns.Score == maxScore {
//			cntOfMaxScore++
//			if rand.Intn(cntOfMaxScore) == 0 {
//				// Replace the candidate with probability of 1/cntOfMaxScore
//				selected = ns.Name
//			}
//		}
//	}
//	return selected, nil
//}
//
//// numFeasibleNodesToFind returns the number of feasible nodes that once found, the scheduler stops
//// its search for more feasible nodes.
//func (g *genericScheduler) numFeasibleNodesToFind(numAllNodes int32) (numNodes int32) {
//	if numAllNodes < minFeasibleNodesToFind || g.percentageOfNodesToScore >= 100 {
//		return numAllNodes
//	}
//
//	adaptivePercentage := g.percentageOfNodesToScore
//	if adaptivePercentage <= 0 {
//		basePercentageOfNodesToScore := int32(50)
//		adaptivePercentage = basePercentageOfNodesToScore - numAllNodes/125
//		if adaptivePercentage < minFeasibleNodesPercentageToFind {
//			adaptivePercentage = minFeasibleNodesPercentageToFind
//		}
//	}
//
//	numNodes = numAllNodes * adaptivePercentage / 100
//	if numNodes < minFeasibleNodesToFind {
//		return minFeasibleNodesToFind
//	}
//
//	return numNodes
//}
//
//// Filters the nodes to find the ones that fit the pod based on the framework
//// filter plugins and filter extenders.
//func (g *genericScheduler) findNodesThatFitPod(ctx context.Context, fwk framework.Framework, state *framework.CycleState, pod *v1.Pod) ([]*v1.Node, framework.NodeToStatusMap, error) {
//	filteredNodesStatuses := make(framework.NodeToStatusMap)
//
//	// Run "prefilter" plugins.
//	_, s := fwk.RunPreFilterPlugins(ctx, state, pod)
//	if !s.IsSuccess() {
//		if !s.IsUnschedulable() {
//			return nil, nil, s.AsError()
//		}
//		// All nodes will have the same status. Some non trivial refactoring is
//		// needed to avoid this copy.
//		allNodes, err := g.nodeInfoSnapshot.NodeInfos().List()
//		if err != nil {
//			return nil, nil, err
//		}
//		for _, n := range allNodes {
//			filteredNodesStatuses[n.Node().Name] = s
//		}
//		return nil, filteredNodesStatuses, nil
//	}
//
//	feasibleNodes, err := g.findNodesThatPassFilters(ctx, fwk, state, pod, filteredNodesStatuses)
//	if err != nil {
//		return nil, nil, err
//	}
//
//	feasibleNodes, err = g.findNodesThatPassExtenders(pod, feasibleNodes, filteredNodesStatuses)
//	if err != nil {
//		return nil, nil, err
//	}
//	return feasibleNodes, filteredNodesStatuses, nil
//}
//
//// findNodesThatPassFilters finds the nodes that fit the filter plugins.
//func (g *genericScheduler) findNodesThatPassFilters(ctx context.Context, fwk framework.Framework, state *framework.CycleState, pod *v1.Pod, statuses framework.NodeToStatusMap) ([]*v1.Node, error) {
//	allNodes, err := g.nodeInfoSnapshot.NodeInfos().List()
//	if err != nil {
//		return nil, err
//	}
//
//	numNodesToFind := g.numFeasibleNodesToFind(int32(len(allNodes)))
//	// Create feasible list with enough space to avoid growing it
//	// and allow assigning.
//	feasibleNodes := make([]*v1.Node, numNodesToFind)
//	if !fwk.HasFilterPlugins() {
//		length := len(allNodes)
//		for i := range feasibleNodes {
//			feasibleNodes[i] = allNodes[(g.nextStartNodeIndex+i)%length].Node()
//		}
//		g.nextStartNodeIndex = (g.nextStartNodeIndex + len(feasibleNodes)) % length
//		return feasibleNodes, nil
//	}
//	return feasibleNodes, nil
//}
//
//// NewGenericScheduler creates a genericScheduler object.
//func NewGenericScheduler(
//	cache internalcache.Cache,
//	nodeInfoSnapshot *internalcache.Snapshot,
//	extenders []framework.Extender,
//	percentageOfNodesToScore int32) ScheduleAlgorithm {
//	return &genericScheduler{
//		cache:                    cache,
//		nodeInfoSnapshot:         nodeInfoSnapshot,
//		percentageOfNodesToScore: percentageOfNodesToScore,
//	}
//}
