# Go Checkpoint Client

[Checkpoint](http://checkpoint.soloio.com) is an internal service at
Solo.io that we use to check version information, broadcast security
bulletins, etc.

We understand that software making remote calls over the internet
for any reason can be undesirable. Because of this, Checkpoint can be
disabled in all of our software that includes it. You can view the source
of this client to see that we're not sending any private information.

Each Solo.io application has it's specific configuration option
to disable checkpoint calls, but the `CHECKPOINT_DISABLE` makes
the underlying checkpoint component itself disabled. For example
in the case of thetool:
```
CHECKPOINT_DISABLE=1 thetool build  all
```

**Note:** This repository is probably useless outside of internal Solo.io
use. It is open source for disclosure and because our open source projects
must be able to link to it.
