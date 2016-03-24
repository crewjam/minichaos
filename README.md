# minichaos

Minichaos is a tool that randomly selects an instance from an autoscaling group and terminates it. If any instance in the group is unhealthy or not marked as "InService" then nothing is terminated.

# Usage

~~~
minichaos --help
Usage of minichaos:
  -asg string
        The name of the autoscaling group. If not specified, the ASG of the currently running instance is used.
  -dry-run
        If true don't actually terminate anything, just pretend to
~~~

# Examples

~~~console
$ minichaos -dry-run -asg barexamplecom-MasterAutoscale-WBKOIE4H3BX7
2016/03/24 14:19:23 i-64a5d0bc: terminated (dry run)
~~~

~~~console
$ minichaos -asg barexamplecom-MasterAutoscale-WBKOIE4H3BX7
2016/03/24 14:19:16 i-1dc3fac7: terminated
~~~

~~~console
$ minichaos -dry-run -asg barexamplecom-MasterAutoscale-WBKOIE4H3BX7
2016/03/24 14:20:23 error: chaos aborted because i-1dc3fac7 health is Unhealthy
~~~
