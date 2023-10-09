# Dagger Docker Module

Docker module provides DinD using Dagger.

## Limitations

This module requires to run Docker service with `InsecureRootCapabilities` enabled. This means that container started 
with `--privileged` flag. This is a security risk and should be used with caution.