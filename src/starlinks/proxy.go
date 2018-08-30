package starlinks

type CachedStorage struct {
	cache CacheStorage
	link  LinkStorage
}

//type CascadeCache struct {
//}

func NewCachedStorage(cache CacheStorage, storage LinkStorage) *CachedStorage {
	return &CachedStorage{
		cache: cache,
		link:  storage,
	}
}

func (cs *CachedStorage) QueryLink(id LinkID) (string, error) {
	link, err := cs.cache.QueryLink(id)
	if err != nil {
		return "", err
	}
	if "" == link {
		link, err = cs.link.QueryLink(id)
		if err != nil {
			return "", err
		}
		if link != "" {
			err = cs.cache.AddLink(id, link)
		}
	}
	return link, err
}

func (cs *CachedStorage) QueryLinks(ids []LinkID) ([]string, error) {
	var links, add_links []string
	var err error

	links, err = cs.cache.QueryLinks(ids)
	if err != nil {
		return nil, err
	}

	missing := make([]uint, 10)
	for i, link := range links {
		if link == "" {
			missing = append(missing, uint(i))
		}
	}
	if len(missing) > 0 {
		missing_ids := make([]LinkID, len(missing))
		for _, pos := range missing {
			missing_ids[pos] = ids[pos]
		}
		add_links, err = cs.link.QueryLinks(missing_ids)
		if err == nil {
			for i, pos := range missing {
				links[pos] = add_links[i]
			}
		}
	}
	return links, err
}

func (cs *CachedStorage) AddLink(url string) (LinkID, error) {
	id, err := cs.link.AddLink(url)
	if err != nil {
		return 0, err
	}
	err = cs.cache.AddLink(id, url)
	return id, err
}

func (cs *CachedStorage) AddLinks(urls []string) ([]LinkID, error) {
	ids, err := cs.link.AddLinks(urls)
	if err != nil {
		return nil, err
	}
	link_map := make(map[LinkID]string, len(ids))
	err = cs.cache.AddLinks(link_map)
	return ids, err
}

func (cs *CachedStorage) RemoveLink(id LinkID) error {
	err := cs.cache.RemoveLink(id)
	if err != nil {
		return err
	}
	err = cs.link.RemoveLink(id)
	return err
}

func (cs *CachedStorage) RemoveLinks(ids []LinkID) error {
	err := cs.cache.RemoveLinks(ids)
	if err != nil {
		return err
	}
	err = cs.link.RemoveLinks(ids)
	return err
}
