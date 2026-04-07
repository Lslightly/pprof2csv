package analyzer

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/pprof/profile"
)

// TrieNode represents a node in a call tree (Trie structure)
type TrieNode struct {
	name     string
	children map[string]*TrieNode
	time     int64
	parent   *TrieNode
}

// CallTree represents the root of a call tree
type CallTree struct {
	root *TrieNode
}

// NewCallTree creates a new empty call tree
func NewCallTree() *CallTree {
	return &CallTree{
		root: &TrieNode{
			children: make(map[string]*TrieNode),
		},
	}
}

// GetOrCreateChild gets or creates a child node
func (n *TrieNode) GetOrCreateChild(name string) *TrieNode {
	if n.children == nil {
		n.children = make(map[string]*TrieNode)
	}

	if child, exists := n.children[name]; exists {
		return child
	}

	child := &TrieNode{
		name:     name,
		children: make(map[string]*TrieNode),
		parent:   n,
	}
	n.children[name] = child
	return child
}

// AddCallChain adds a call chain to the tree and accumulates time
// chain is from caller to target (e.g., ["A", "B", "mallocgc"])
func (t *CallTree) AddCallChain(chain []string, time int64) *TrieNode {
	current := t.root
	for _, name := range chain {
		current = current.GetOrCreateChild(name)
	}
	current.time += time
	return current
}

// GetCallChain retrieves the call chain from root(leaf) to target
func (leaf *TrieNode) GetCallChain() []string {
	var chain []string
	for node := leaf; node != nil && node.name != ""; node = node.parent {
		chain = append(chain, node.name)
	}
	return chain
}

// String converts call chain to string representation
func (leaf *TrieNode) String() string {
	chain := leaf.GetCallChain()
	if len(chain) == 0 {
		return ""
	}
	return strings.Join(chain, " → ")
}

// traverseTreeForLeaves traverses the tree and calls visitor for each leaf node
func traverseTreeForLeaves(node *TrieNode, visitor func(*TrieNode)) {
	if len(node.children) == 0 {
		visitor(node)
		return
	}

	for _, child := range node.children {
		traverseTreeForLeaves(child, visitor)
	}
}

// GetCallerKNameSet retrieves the set of k-hop caller function names for a given callee.
//
// Parameters:
//   - filename: path to the pprof profile file
//   - callee: the target function name to find in the call stack
//   - k: the number of hops up the call stack to find the caller (k=1 for direct caller)
//   - showFrom: if non-empty, only samples whose stacktrace contains this function are included
//
// Returns:
//   - []string: unique set of k-hop caller function names
//   - error: an error if the callee is not found in the profile or if the profile cannot be parsed
//
// The function searches through all sample call stacks in the profile.
// If showFrom is non-empty, only samples containing showFrom in their call stack are considered.
// For each sample, it finds the callee in the call stack and then returns the function
// name that is k positions above it in the stack. The result is deduplicated before returning.
//
// Example:
//
//	If the call stack is: [main] -> [foo] -> [bar] -> [baz] (where baz is the leaf)
//	- GetCallerKNameSet("profile.pprof", "baz", 1, "") returns ["bar"]
//	- GetCallerKNameSet("profile.pprof", "baz", 2, "") returns ["foo"]
//	- GetCallerKNameSet("profile.pprof", "baz", 3, "") returns ["main"]
func GetCallerKNameSet(filename string, callee string, k int, showFrom string) (result []string, err error) {
	// Load and parse profile data
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error loading profile: %v", err)
	}

	p, err := profile.ParseData(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse profile data: %w", err)
	}

	callerSet := make(map[string]struct{})
	calleeFound := false

	// Process each sample's call stack
	for _, sample := range p.Sample {
		// Filter: skip sample if showFrom specified but not found in stacktrace
		if showFrom != "" {
			found := false
		locationLoop:
			for _, loc := range sample.Location {
				for _, le := range loc.Line {
					if le.Function != nil && le.Function.Name == showFrom {
						found = true
						break locationLoop
					}
				}
			}
			if !found {
				continue
			}
		}

		// Search for callee in the call stack
		for calleeIdx, loc := range sample.Location {
			// Skip locations without lines
			if len(loc.Line) == 0 {
				continue
			}

			// Check if any line in this location matches the callee
			isCallee := false
			for _, lineEntry := range loc.Line {
				if lineEntry.Function != nil && lineEntry.Function.Name == callee {
					isCallee = true
					break
				}
			}

			if isCallee {
				calleeFound = true
				// Calculate the index of the k-hop caller
				callerIdx := calleeIdx + k

				// Check if caller index is within valid range
				if callerIdx >= 0 && callerIdx < len(sample.Location) {
					callerLoc := sample.Location[callerIdx]

					// Get function name from caller location
					for _, lineEntry := range callerLoc.Line {
						if lineEntry.Function != nil && lineEntry.Function.Name != "" {
							callerSet[lineEntry.Function.Name] = struct{}{}
							break // Only need one function name per location
						}
					}
				}
				break // Found callee in this sample, move to next sample
			}
		}
	}

	// Return error if callee was never found
	if !calleeFound {
		return nil, fmt.Errorf("callee function '%s' not found in the profile", callee)
	}

	// Convert map to slice
	result = make([]string, 0, len(callerSet))
	for name := range callerSet {
		result = append(result, name)
	}

	return result, nil
}

// GetCalleeKNameSet retrieves the set of k-hop callee function names for a given caller.
//
// Parameters:
//   - filename: path to the pprof profile file
//   - caller: the target function name to find in the call stack
//   - k: the number of hops down the call stack to find the callee (k=1 for direct callee)
//   - showFrom: if non-empty, only samples whose stacktrace contains this function are included
//
// Returns:
//   - []string: unique set of k-hop callee function names
//   - error: an error if the caller is not found in the profile or if the profile cannot be parsed
//
// The function searches through all sample call stacks in the profile.
// If showFrom is non-empty, only samples containing showFrom in their call stack are considered.
// For each sample, it finds the caller in the call stack and then returns the function
// name that is k positions below it in the stack. The result is deduplicated before returning.
//
// Example:
//
//	If the call stack is: [main] -> [foo] -> [bar] -> [baz] (where baz is the leaf)
//	- GetCalleeKNameSet("profile.pprof", "main", 1, "") returns ["foo"]
//	- GetCalleeKNameSet("profile.pprof", "main", 2, "") returns ["bar"]
//	- GetCalleeKNameSet("profile.pprof", "main", 3, "") returns ["baz"]
func GetCalleeKNameSet(filename string, caller string, k int, showFrom string) (result []string, err error) {
	// Load and parse profile data
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error loading profile: %v", err)
	}

	p, err := profile.ParseData(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse profile data: %w", err)
	}

	calleeSet := make(map[string]struct{})
	callerFound := false

	// Process each sample's call stack
	for _, sample := range p.Sample {
		// Filter: skip sample if showFrom specified but not found in stacktrace
		if showFrom != "" {
			found := false
		locationLoop:
			for _, loc := range sample.Location {
				for _, le := range loc.Line {
					if le.Function != nil && le.Function.Name == showFrom {
						found = true
						break locationLoop
					}
				}
			}
			if !found {
				continue
			}
		}

		// Search for caller in the call stack
		for callerIdx, loc := range sample.Location {
			// Skip locations without lines
			if len(loc.Line) == 0 {
				continue
			}

			// Check if any line in this location matches the caller
			isCaller := false
			for _, lineEntry := range loc.Line {
				if lineEntry.Function != nil && lineEntry.Function.Name == caller {
					isCaller = true
					break
				}
			}

			if isCaller {
				callerFound = true
				// Calculate the index of the k-hop callee
				calleeIdx := callerIdx - k

				// Check if callee index is within valid range
				if calleeIdx >= 0 && calleeIdx < len(sample.Location) {
					calleeLoc := sample.Location[calleeIdx]

					// Get function name from callee location
					for _, lineEntry := range calleeLoc.Line {
						if lineEntry.Function != nil && lineEntry.Function.Name != "" {
							calleeSet[lineEntry.Function.Name] = struct{}{}
							break // Only need one function name per location
						}
					}
				}
				break // Found caller in this sample, move to next sample
			}
		}
	}

	// Return error if caller was never found
	if !callerFound {
		return nil, fmt.Errorf("caller function '%s' not found in the profile", caller)
	}

	// Convert map to slice
	result = make([]string, 0, len(calleeSet))
	for name := range calleeSet {
		result = append(result, name)
	}

	return result, nil
}

// GetCallerPercentage returns the percentage of total target function time attributed to each caller.
//
// This function is a convenience wrapper around GetCallerChainPercentageWithTree with k=1,
// which analyzes direct (1-hop) callers of the target function.
//
// Parameters:
//   - filename: path to the pprof profile file
//   - target: the target function name to analyze (e.g., "runtime.mallocgc")
//   - showFrom: if non-empty, only samples whose stacktrace contains this function are included
//
// Returns:
//   - map[string]float64: map of caller function name to percentage (0-100)
//   - totalDuration: total duration of the target function
//   - error: if the target is not found or profile cannot be parsed
//
// Example:
//
//	percentages, total, err := GetCallerPercentage("profile.pprof", "runtime.mallocgc", "")
//	// percentages might be: {"runtime.newobject": 45.2, "runtime.makeSlice": 30.5, ...}
func GetCallerPercentage(filename string, target string, showFrom string) (map[string]float64, time.Duration, error) {
	// Use the more general GetCallerChainPercentageWithTree with k=1 for direct callers
	tree, leafPercentages, totalDuration, err := GetCallerChainPercentageWithTree(filename, target, 1, showFrom)
	if err != nil {
		return nil, 0, err
	}

	// Extract caller percentages from the tree structure
	// When k=1, the tree structure is: root → caller → target (leaf)
	callerPercentages := extractCallerPercentagesFromTree(tree, leafPercentages)

	return callerPercentages, totalDuration, nil
}

// GetCallerChainPercentageWithTree returns the call tree with accumulated times
// and a map of leaf nodes to their percentages.
//
// Parameters:
//   - filename: path to the pprof profile file
//   - target: target function name to analyze (e.g., "runtime.mallocgc")
//   - k: number of hops up the call stack (k=1 for direct caller)
//   - showFrom: if non-empty, only samples whose stacktrace contains this function are included
//
// Returns:
//   - *CallTree: the call tree structure
//   - map[*TrieNode]float64: map of leaf node pointers to percentage (0-100)
//   - time.Duration: total duration of target function
//   - error: if target is not found or profile cannot be parsed
func GetCallerChainPercentageWithTree(filename string, target string, k int, showFrom string) (*CallTree, map[*TrieNode]float64, time.Duration, error) {
	// Load and parse profile data
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("error loading profile: %v", err)
	}

	p, err := profile.ParseData(data)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to parse profile data: %w", err)
	}

	tree := NewCallTree()
	var totalTargetTime int64
	targetFound := false
	timeUnit := convTimeUnit(p.SampleType[1].Unit)

	// Process each sample's call stack
	for _, sample := range p.Sample {
		// Get the value (time) for this sample
		var value int64
		if len(sample.Value) > 1 {
			value = sample.Value[1]
		}

		// Filter: skip sample if showFrom specified but not found in stacktrace
		if showFrom != "" {
			found := false
		locationLoop:
			for _, loc := range sample.Location {
				for _, le := range loc.Line {
					if le.Function != nil && le.Function.Name == showFrom {
						found = true
						break locationLoop
					}
				}
			}
			if !found {
				continue
			}
		}

		// Search for target in the call stack
		for targetIdx, loc := range sample.Location {
			// Skip locations without lines
			if len(loc.Line) == 0 {
				continue
			}

			// Check if any line in this location matches the target
			isTarget := false
			for _, lineEntry := range loc.Line {
				if lineEntry.Function != nil && lineEntry.Function.Name == target {
					isTarget = true
					break
				}
			}

			if !isTarget {
				continue
			}
			targetFound = true
			totalTargetTime += value

			// Build call chain slice from k-hop caller
			chain := buildCallChainSlice(sample.Location, targetIdx, k)

			// Add to tree (returns leaf node pointer)
			tree.AddCallChain(chain, value)

			break // Found target in this sample, move to next sample
		}
	}

	// Return error if target was never found
	if !targetFound {
		return nil, nil, 0, fmt.Errorf("target function '%s' not found in the profile", target)
	}

	// Calculate percentages for each leaf node
	leafPercentages := make(map[*TrieNode]float64)
	traverseTreeForLeaves(tree.root, func(leaf *TrieNode) {
		leafPercentages[leaf] = float64(leaf.time) / float64(totalTargetTime) * 100.0
	})

	return tree, leafPercentages, time.Duration(totalTargetTime) * timeUnit, nil
}

// buildCallChainSlice builds a call chain slice from k-hop caller to target.
// The chain always starts with the target function (locations[targetIdx]).
// If a location doesn't have a valid function name, it's skipped from the chain.
// This may result in a shorter chain if profile data is incomplete.
func buildCallChainSlice(locations []*profile.Location, targetIdx, k int) []string {
	chain := []string{}

	// From target to k-hop caller
	for i := targetIdx; i <= targetIdx+k; i++ {
		if i >= 0 && i < len(locations) {
			loc := locations[i]
			if len(loc.Line) > 0 {
				// Find the first valid function name in this location
				// If no valid function name is found, skip this location
				if funcName := getFunctionNameFromLocation(loc); funcName != "" {
					chain = append(chain, funcName)
				}
			}
		}
	}

	return chain
}

// getFunctionNameFromLocation extracts the first valid function name from a location.
// Searches through all Line entries and returns the first non-empty function name.
// Returns empty string if no valid function name is found.
func getFunctionNameFromLocation(loc *profile.Location) string {
	for _, line := range loc.Line {
		if line.Function != nil && line.Function.Name != "" {
			return line.Function.Name
		}
	}
	return ""
}

// extractCallerPercentagesFromTree extracts caller percentages from a call tree.
// When k=1, the tree structure is: root -> target -> caller(leaf)
// This function extracts percentages for direct children of target node (callers).
func extractCallerPercentagesFromTree(tree *CallTree, leafPercentages map[*TrieNode]float64) map[string]float64 {
	callerPercentages := make(map[string]float64)

	// Find the target node (first child of root, since all paths start with target)
	var targetNode *TrieNode
	for _, child := range tree.root.children {
		targetNode = child
		break // All paths start with the same target function
	}

	if targetNode == nil {
		return callerPercentages
	}

	// For k=1, the caller nodes are the direct children of targetNode
	// Since k=1, caller nodes are themselves leaf nodes
	for callerName, callerNode := range targetNode.children {
		callerPercentages[callerName] += leafPercentages[callerNode]
	}

	return callerPercentages
}
