# oci-hooks-mount-chown
An OCI hook for changing owner for a mount point

# Why

Some container runtime tools like podman allows you to mount image but it doesn't provide an option to change the owner of the mounted file system.
As a result, this limits what you can do with the mounted volume.
A podman issue ([#18986](https://github.com/containers/podman/issues/18986)) has been opened for this particular need.
Before that issue is closed, or say if one needs to chown for any mount point inside a container, this hook comes hady.

# How

To use this hook for changing the own of a mount point, there are a few special annotations you can add to the container:

- com.launchplatform.oci-hooks.mount-chown.**&lt;NAME&gt;**.path
- com.launchplatform.oci-hooks.mount-chown.**&lt;NAME&gt;**.owner
- com.launchplatform.oci-hooks.mount-chown.**&lt;NAME&gt;**.policy
- com.launchplatform.oci-hooks.mount-chown.**&lt;NAME&gt;**.mode

The `NAME` can be any valid annotation string without a dot in it.
The `path` and `owner` annotations with the same name need to appear in pairs, otherwise it will be ignored.
The owner value can be a single uid integer value or uid plus gid, with a format like `UID[:GID]`.
Please note that username is not supported, only integer value works.
The `policy` annoation is optional, there are two available options:

- `recursive` - chown recursively (default)
- `root-only` - chown only for the root folder of mount-pooint

If the policy value is not provided, `recursive` will be used by default.
With these annotations, to change owner of a path, here's an example of podman command you can run:

```bash
podman run \
    --user 2000:2000 \
    --annotation=com.launchplatform.oci-hooks.mount-chown.data.path=/data \
    --annotation=com.launchplatform.oci-hooks.mount-chown.data.owner=2000:2000 \
    --annotation=com.launchplatform.oci-hooks.mount-chown.data.policy=root-only \
    --mount type=image,source=my-data-image,destination=/data,rw=true \
    -it alpine
# Now you can write to the root folder of the image mount
touch /data/my-data.lock
```

The `mode` option can also be used. However, please note that it only changes the mode of root path, it doesn't apply recursively regardless what the `policy` says.
Either one of `owner` or `mode` needs to be provided.
For now podman's image mount comes with `0555` as the root folder, without changing the owner, changing the mode to `0777` might help.
Here's an example:

```bash
podman run \
    --user 2000:2000 \
    --annotation=com.launchplatform.oci-hooks.mount-chown.data.path=/data \
    --annotation=com.launchplatform.oci-hooks.mount-chown.data.mode=777 \
    --mount type=image,source=my-data-image,destination=/data,rw=true \
    -it alpine
# Now you can write to the root folder of the image mount
touch /data/my-data.lock
```

## Add createContainer hook directly in the OCI spec

There are different ways of running a container, if you are generating OCI spec yourself and running OCI runtimes such as [crun](https://github.com/containers/crun) yourself, you can add the `createContainer` hook directly into the spec file like this:

```json
{
  "//": "... other OCI spec content ...",
  "hooks": {
    "createContainer": [
      {
        "path": "/usr/bin/mount_chown"
      }
    ]
  }
}
```

For more information about the OCI spec schema, please see the [document here](https://github.com/opencontainers/runtime-spec/blob/48415de180cf7d5168ca53a5aa27b6fcec8e4d81/config.md#posix-platform-hooks).

## Add OCI hook config

Another way to add the OCI hook is to create a OCI hook config file.
Here's an example:

```json
{
  "version": "1.0.0",
  "hook": {
    "path": "/usr/bin/mount_chown"
  },
  "when": {
    "annotations": {
        "com\\.launchplatform\\.oci-hooks\\.mount-chown\\.([^.]+)\\.path": "(.+)"
    }
  },
  "stages": ["createContainer"]
}
```

For more information about the OCI hooks schema, please see the [document here](https://github.com/containers/podman/blob/v3.4.7/pkg/hooks/docs/oci-hooks.5.md).

# Debug

To debug the hook, you can add `--log-level=debug` (or `trace` if you need more details) argument for the `archive_overlay` executable, it will print debug information.
With OCI runtimes like [crun](https://github.com/containers/crun), you can also add an annotation like this:

```
run.oci.hooks.stderr=/path/to/stderr
```

to make the runtime redirect the stderr from the hook executable to specific file.
Please note that podman invokes poststop hook instead of delegating it to crun, so the annotation won't work for podman.

