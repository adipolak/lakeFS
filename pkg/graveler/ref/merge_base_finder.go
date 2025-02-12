package ref

import (
	"container/heap"
	"context"

	"github.com/treeverse/lakefs/pkg/graveler"
)

type CommitGetter interface {
	GetCommit(ctx context.Context, repositoryID graveler.RepositoryID, commitID graveler.CommitID) (*graveler.Commit, error)
}

type reachedFlags uint8

const (
	fromLeft reachedFlags = 1 << iota
	fromRight
)

// FindMergeBase finds the best common ancestor according to the definition in the git-merge-base documentation: https://git-scm.com/docs/git-merge-base
// One common ancestor is better than another common ancestor if the latter is an ancestor of the former.
func FindMergeBase(ctx context.Context, getter CommitGetter, repositoryID graveler.RepositoryID, leftID, rightID graveler.CommitID) (*graveler.Commit, error) {
	var commitRecord *graveler.CommitRecord
	queue := NewCommitsGenerationPriorityQueue()
	reached := make(map[graveler.CommitID]reachedFlags)
	reached[rightID] |= fromRight
	reached[leftID] |= fromLeft
	commit, err := getCommitAndEnqueue(ctx, getter, &queue, repositoryID, leftID)
	if err != nil {
		return nil, err
	}
	if leftID == rightID {
		return commit, nil
	}

	_, err = getCommitAndEnqueue(ctx, getter, &queue, repositoryID, rightID)
	if err != nil {
		return nil, err
	}
	for {
		if queue.Len() == 0 {
			return nil, nil
		}
		commitRecord = heap.Pop(&queue).(*graveler.CommitRecord)
		commitFlags := reached[commitRecord.CommitID]
		for _, parent := range commitRecord.Parents {
			if _, exist := reached[parent]; !exist {
				// parent commit is queued only if it was not handled before. Otherwise it, and
				// all its ancestors were already queued and so, will have entries in 'reached' map
				_, err := getCommitAndEnqueue(ctx, getter, &queue, repositoryID, parent)
				if err != nil {
					return nil, err
				}
			}
			// mark the parent with the flag values from its descendents. This is done regardless
			// of whether this parent commit is being queued in the current iteration or not. In
			// both cases, if the 'reached' update signifies it was reached from both left and
			// right nodes - it is the requested parent node
			reached[parent] |= commitFlags
			if reached[parent]&fromLeft != 0 && reached[parent]&fromRight != 0 {
				// commit was reached from both left and right nodes
				return getter.GetCommit(ctx, repositoryID, parent)
			}
		}
	}
}

func getCommitAndEnqueue(ctx context.Context, getter CommitGetter, queue *CommitsGenerationPriorityQueue, repositoryID graveler.RepositoryID, commitID graveler.CommitID) (*graveler.Commit, error) {
	commit, err := getter.GetCommit(ctx, repositoryID, commitID)
	if err != nil {
		return nil, err
	}
	heap.Push(queue, &graveler.CommitRecord{CommitID: commitID, Commit: commit})
	return commit, nil
}
