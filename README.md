# oci-hooks-mount-chown
An OCI hook for changing owner for a mount point

# Why

Some container runtime tools like podman allows you to mount image but it doesn't provide an option to change the owner of the mounted file system.
As a result, this limits what you can do with the mounted volume.
A podman issue ([#18986](https://github.com/containers/podman/issues/18986)) has been opened for this particular need.
Before that issue is closed, or say if one needs to chown for any mount point inside a container, this hook comes hady.

# How

To use this hook for changing the own of a mount point, there are a few special annotations you can add to the container:

- com.launchplatform.oci-hooks.mount-chown.<NAME>.mount-point
- com.launchplatform.oci-hooks.mount-chown.<NAME>.owner
- com.launchplatform.oci-hooks.mount-chown.<NAME>.policy (optional)

The `NAME` can be any valid annotation string without a dot in it.
The `mount-point` and `owner` annotations with the same name need to appear in pairs, otherwise it will be ignored.
The owner value can be a single uid integer value or uid plus gid, with a format like `UID[:GID]`.
Please note that username is not supported, only integer value works.
The `policy` annoation is optional, there are two available options:

- `recursive` - chown recursively (default)
- `root-only` - chown only for the root folder of mount-pooint

If the policy value is not provided, `recursive` will be used by default.
With these annotations, to change owner of a mount-point, here's an example of podman command you can run:

```bash
podman run \
    --user 2000:2000 \
    --annotation=com.launchplatform.oci-hooks.mount-chown.data.mount-point=/data \
    --annotation=com.launchplatform.oci-hooks.mount-chowny.data.owner=2000:2000 \
    --mount type=image,source=my-data-image,destination=/data,rw=true \
    -it alpine
# Now you can write to the root folder of the image mount
touch /data/my-data.lock
```
