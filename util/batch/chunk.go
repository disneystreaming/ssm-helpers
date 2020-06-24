package batch

// ChunkFunc is a function when used e.g with chunk, provides a minimum and maximum range
// to batch the items of a slice.
type ChunkFunc func(min int, max int) (bool, error)

func Chunk(length, batch int, fn ChunkFunc) error {
	for i := 0; i < length; i += batch {
		j := i + batch

		if j > length {
			j = length
		}

		cont, err := fn(i, j)

		if err != nil {
			return err
		}

		if !cont {
			return nil
		}
	}

	return nil
}
