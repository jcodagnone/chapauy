// Copyright 2025 The ChapaUY Authors
// SPDX-License-Identifier: Apache-2.0

package curation

// clusterJudgments groups judgments into clusters based on a distance threshold.
func clusterJudgments(judgments []*Location, distanceThreshold float64) [][]*Location {
	clusters := make([][]*Location, 0, len(judgments))

	visited := make([]bool, len(judgments))

	for i, j1 := range judgments {
		if visited[i] {
			continue
		}

		cluster := []*Location{j1}
		visited[i] = true

		for j, j2 := range judgments {
			if visited[j] {
				continue
			}

			// Check distance against all members of the current cluster
			for _, member := range cluster {
				if j2.Point.HaversineDistance(member.Point) <= distanceThreshold {
					cluster = append(cluster, j2)
					visited[j] = true

					break // Move to next judgment once it's added to the cluster
				}
			}
		}

		clusters = append(clusters, cluster)
	}

	return clusters
}
