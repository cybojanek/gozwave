// Package cache caches node information in a local directory.
package cache

import (
	"bytes"
	"fmt"
	"github.com/cybojanek/gozwave/network"
	"io/ioutil"
	"os"
	"path"
)

type NodeCache struct {
	Directory string
}

// LoadNodes loads all the nodes from the cache/network, and updates the cache.
func (cache *NodeCache) LoadNodes(net *network.Network) error {
	// Crate cache directory if it does not exist.
	if err := os.Mkdir(cache.Directory, 0755); err != nil && !os.IsExist(err) {
		return err
	}

	// Iterate over nodes and refresh them, querying for more specific
	// information about their supported operations
	for _, n := range net.GetNodes() {

		// Read existing cache file.
		cacheFile := path.Join(cache.Directory, fmt.Sprintf("%d.json", n.ID))
		cacheBytes, err := ioutil.ReadFile(cacheFile)
		if err != nil {
			cacheBytes = nil
			if !os.IsNotExist(err) {
				return err
			}
		}

		// Load node and get new cache bytes.
		newCacheBytes, err := n.Load(cacheBytes)
		if err != nil {
			return err
		}

		// If we have new cache bytes, and they are different, then update
		// cache.
		if newCacheBytes != nil && !bytes.Equal(cacheBytes, newCacheBytes) {
			if err := ioutil.WriteFile(cacheFile, newCacheBytes, 0644); err != nil {
				return err
			}
		}
	}

	return nil
}
