# ssm session

Start a session with a single AWS instance using `--instances`.

![](../../img/ssm-session-1.gif)

Start a session with multiple instances across `--regions` and `--profiles` based on instance tags using `--filter`, or use the interactive prompt to select instances on the fly.

Note: multiple regions and accounts takes longer for API calls.

![](../../img/ssm-session-2.gif)

## about

`ssm session` is a tool for finding and connecting to SSM-managed EC2 instances running Linux and the SSM Agent. It uses the Amazon-supplied `session-manager-plugin` binary in combination with the AWS CLI tool to create the actual sessions. If multiple instances are specified or selected, the sessions will be multiplexed in a `tmux` session, and the user will be dropped into the session before the tool exits.

### basic usage

By default, `ssm session` will attempt to search/connect for instances in your default AWS profile (specified in your ~/.aws/config, or via the `AWS_PROFILE` environment variable if set), and the default region linked to that profile (or the `AWS_REGION` environment variable, if set).

The order of preference for profile/region settings is set as follows:

1) explicitly specified by the --profiles and/or --regions flags
2) explicitly set using the `AWS_PROFILE` and/or `AWS_REGION` environment variables
3) falling back to the default profile and/or region as specified in your ~/.aws/config file.

To use a specific profile and/or region, use the `-p (--profiles)` or `-r (--regions)` flags.

e.g. `-p account1,account2 -r us-east-1,us-west-2`

Take careful note that if multiple regions *and* profiles are specified, `ssm-session` will return results for instances across all the possible permutations.

For example, using the flags `--profiles foo,bar,baz --regions us-east-1,us-west-2,eu-east-1` will target instances in each of the profile/region combinations:

    foo@us-east-1, foo@us-west-2, foo@eu-east-1
    bar@us-east-1, bar@us-west-2, bar@eu-east-1
    baz@us-east-1, baz@us-west-2, baz@eu-east-1

As such, please be careful when doing so and when using the `--all-profiles` flag.

If a region is not specified, you will receive the following warning, but execution will continue:

```
WARNING No AWS region setting found. Will default to the region linked to any profile in use.
```

Any number of tags can be specified using multiple `-t` flags. The instance data for that tag will be displayed as an additional column in the selection dialog if multiple instances are found. Columns for Instance ID, Region, and Profile are always displayed.

#### connecting to a single instance

`ssm session -i i-12345`

#### connecting to multiple instances

`ssm session -i i-12345,i-67890`

#### searching for instances

By default, if no instance or filters are specified, `ssm session` will target all instances in the current account + region.

```
> ssm session

INFO    Retrieved 7 usable instances.
       Instance ID   Region Profile
? Select the instances to which you want to connect:  [Use arrows to move, space to select, type to filter]
> [ ]  i-0d770cb     us-east-1  profile1
  [ ]  i-0b75ad5     us-east-1  profile1
  [ ]  i-0fb5f18     us-east-1  profile1
  [ ]  i-01e6823     us-east-1  profile1
  [ ]  i-0ab2a85     us-east-1  profile1
  [ ]  i-0d3558e     us-east-1  profile1
  [ ]  i-026a8f0     us-east-1  profile1
```

#### searching for instances with tag-based filters

Tag-based filtering can also be applied to your search results (including if you manually specify instance names). These filters are additive, which means that each filter you provide will prune down your results to include only instances that match *all* of the provided filters.

```
> ssm session -p profile1 -f env=prod -f app=myapp

INFO    Retrieved 6 usable instances.
       Instance ID  Region     Profile
? Select the instances to which you want to connect:  [Use arrows to move, space to select, type to filter]
> [ ]  i-7268418e   us-east-1  profile1
  [ ]  i-7b6f4687   us-east-1  profile1
  [ ]  i-9cdd364a   us-east-1  profile1
  [ ]  i-ccdc371a   us-east-1  profile1
  [ ]  i-1289eeef   us-east-1  profile1
  [ ]  i-2f96f1d2   us-east-1  profile1
```

If we add those columns to the instance list using the `-t` option, we can see the effect in practice:

```
> ssm session -p profile1 -f env=prod -f app=myapp -t app -t env

INFO    Retrieved 6 usable instances.
       Instance ID  Region     Profile   app    env
? Select the instances to which you want to connect:  [Use arrows to move, space to select, type to filter]:
  [ ]  i-7b6f4687   us-east-1  profile1 myapp  prod
  [ ]  i-7b6f4687   us-east-1  profile1 myapp  prod
  [ ]  i-9cdd364a   us-east-1  profile1 myapp  prod
  [ ]  i-ccdc371a   us-east-1  profile1 myapp  prod
  [ ]  i-1289eeef   us-east-1  profile1 myapp  prod
  [ ]  i-2f96f1d2   us-east-1  profile1 myapp  prod
```

#### searching for instances in multiple accounts and/or regions

```
> ssm session -p profile1 -r us-east-1,us-west-2

INFO    Retrieved 9 usable instances.
       Instance ID          Region     Profile
? Select the instances to which you want to connect:  [Use arrows to move, space to select, type to filter]
> [ ]  i-0d770cb81ae0fc316      us-east-1  profile1
  [ ]  i-0b75ad53689daabf8      us-east-1  profile1
  [ ]  i-0f267dedb9a979fd1      us-west-2  profile1
  [ ]  i-0fb5f186125becc0d      us-east-1  profile1
  [ ]  i-01e6823b89fd224b2      us-east-1  profile1
  [ ]  i-0ab2a8562dda37865      us-east-1  profile1
  [ ]  i-0d3558e702b02e3dc      us-east-1  profile1
  [ ]  i-026a8f0ed1ace92aa      us-east-1  profile1
```

### usage flags

```
    --all-profiles
        [USE WITH CAUTION] Parse through ~/.aws/config to target all profiles.
    --dry-run
        Retrieve the list of profiles, regions, and instances your command(s) would target
    -f, --filter strings
        Filter instances based on tag value. Tags are evaluated with logical AND (instances must match all tags).
        Multiple allowed, delimited by commas (e.g. env=dev,foo=bar)
    -h, --help
        help for session
    -i, --instance strings
        Specify what instance IDs you want to target.
        Multiple allowed, delimited by commas (e.g. --instance i-12345,i-23456)
    -l, --limit int
        Set a limit for the number of instance results returned per profile/region combination. (default 10)
    -p, --profile strings
        Specify a specific profile to use with your API calls.
        Multiple allowed, delimited by commas (e.g. --profile profile1,profile2)
    -r, --region strings
        Specify a specific region to use with your API calls.
        This option will override any profile settings in your config file.
        Multiple allowed, delimited by commas (e.g. --region us-east-1,us-west-2)
                              
        [NOTE] Mixing --profile and --region will result in your command targeting every matching instance in the selected profiles and regions.
        e.g., "--profile foo,bar,baz --region us-east-1,us-west-2,eu-east-1" will target instances in each of the profile/region combinations:
            "foo@us-east-1, foo@us-west-2, foo@eu-east-1"
            "bar@us-east-1, bar@us-west-2, bar@eu-east-1"
            "baz@us-east-1, baz@us-west-2, baz@eu-east-1"
        Please be careful.
    --session-name string
        Specify a name for the tmux session created when multiple instances are selected (default "ssm-session")
    -t, --tag strings
        Adds the specified tag as an additional column to be displayed during the instance selection prompt.
```
