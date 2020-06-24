package blub

import (
	"fmt"
	"github.com/golangee/log"
	"os"
)

type PageStore struct {
	logger   log.Logger
	file     *os.File
	pageSize int64
	buf      *Page
	length   int64
}

func OpenPageStore(fname string, pageSize int64) (*PageStore, error) {
	file, err := os.OpenFile(fname, os.O_APPEND|os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return nil, err
	}

	p := &PageStore{file: file, logger: log.New("pageStore", log.Obj("db", fname))}

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	if stat.Size()%pageSize != 0 {
		return nil, fmt.Errorf("database size is invalid, must be multiple of page size")
	}

	p.pageSize = pageSize
	p.buf = NewPage(int(p.pageSize))
	p.length = stat.Size()

	return p, nil
}

func (p *PageStore) Size() int64 {
	return p.length
}

func (p *PageStore) PageCount() int64 {
	return p.Size() / p.pageSize
}

func (p *PageStore) Close() error {
	return p.file.Close()
}

func (p *PageStore) ForEachPage(f func(page *Page) error) error {
	pages := p.PageCount()
	page := p.buf

	for i := int64(0); i < pages; i++ {
		offset := i * p.pageSize
		n, err := p.file.ReadAt(page.buf, offset)
		if err != nil {
			return fmt.Errorf("unable to read page at %d: %w", offset, err)
		}

		if n != page.Size() {
			return fmt.Errorf("buffer underflow while reading page at %d, expected %d but got %d", offset, page.Size(), n)
		}

		err = f(page)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *PageStore) AddPage(f func(page *Page) error) error {
	page := p.buf
	page.clear()

	err := f(page)
	if err != nil {
		return err
	}

	offset := p.PageCount() * p.pageSize
	n, err := p.file.WriteAt(page.buf, offset)
	if err != nil {
		return fmt.Errorf("unable to read page at %d: %w", offset, err)
	}

	if n != len(page.buf) {
		return fmt.Errorf("buffer underun while writing page at %d, expected %d but got %d", offset, page.Size(), n)
	}

	return nil
}
