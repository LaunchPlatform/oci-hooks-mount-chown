# oci-hooks-mount-chown
An OCI hook for changing owner for a mount point

# Why

Some container runtime tools like podman allows you to mount image but it doesn't provide an option to change the owner of the mounted file system.
As a result, this limits what you can do with the mounted volume.
A podman issue ([#18986](https://github.com/containers/podman/issues/18986)) has been opened for this particular need.
Before that issue is closed, or say if one needs to chown for any mount point inside a container, this hook comes hady.
