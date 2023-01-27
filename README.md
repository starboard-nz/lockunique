## LockUnique

LockUnique implements locking by a unique ID. This is useful to perform tasks for many IDs in parallel but restricting
to one task per ID.

The ID can be any comparable type, LockUnique uses generics.

## Example

```
	l := lockunique.NewLockUnique[int32]()

        id := int32(123)

        l.Lock(id)
        do_something(id)
        err := l.Unlock(id)
        if err != nil {
                // id was not locked
        }
```
