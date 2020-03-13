# ssm-run

Run a command on a list of instances or based on instance tags.
Increase `-log-level` to see individual instances or command output.
Run multiple commands in the same shell context (e.g. `-c 'uptime;hostname'`) or a local script with `-f`.

![](../../img/ssm-run.gif)

## about

`ssm-run` is a tool for executing commands on SSM-managed EC2 instances without having to connect to them directly.

### basic usage

By default, `ssm-run` will attempt to search/connect for instances in your default AWS profile (specified in your ~/.aws/config, or via the `AWS_PROFILE` environment variable if set), and the default region linked to that profile (or the `AWS_REGION` environment variable, if set).

The order of preference for profile/region settings is set as follows:

1) explicitly specified by the --profiles and/or --regions flags
2) explicitly set using the `AWS_PROFILE` and/or `AWS_REGION` environment variables
3) falling back to the default profile and/or region as specified in your ~/.aws/config file.

To use a specific profile and/or region, use the `-p (--profiles)` or `-r (--regions)` flags.

e.g. `-p account1,account2 -r us-east-1,us-west-2`

Take careful note that if multiple regions *and* profiles are specified, `ssm-run` will attempt to run your commands across all the possible permutations.

For example, using the flags `--profiles foo,bar,baz --regions us-east-1,us-west-2,eu-east-1` will target instances in each of the profile/region combinations:

	foo@us-east-1, foo@us-west-2, foo@eu-east-1
	bar@us-east-1, bar@us-west-2, bar@eu-east-1
	baz@us-east-1, baz@us-west-2, baz@eu-east-1

As such, please be careful when doing so and when using the `--all-profiles` flag.

If a region is not specified, you will receive the following warning, but execution will continue:

```
WARNING	No AWS region setting found. Will default to the region linked to any profile in use.
```

The use of the `--limit` flag, if set, will result in your command being executed on the specified number of random instances returned for each profile/region combination. For example, if you set `--limit=10` on the account `foo` for regions `us-east-1` and `us-west-2`, your command would be executed on up to 20 total instances.

#### running command(s) on a single instance

```
> ./ssm-run -i i-12345 -c "uname -a"

INFO    Command(s) to be executed: uname -a
INFO    Fetched 1 instances for account [profile1] in [us-east-1].
INFO    Instance ID              Region          Profile         Status
INFO    i-12345					 us-east-1       profile1         Success
INFO    Execution results: 1 SUCCESS, 0 FAILED
```

Running multiple commands is easy; just separate your commands with semicolons (;). These commands will be executed in series, in the same console context (e.g., variables you set can be used between commands).

```
> ./ssm-run -i i-12345 -c "uname -a; hostname; ifconfig"

INFO    Command(s) to be executed: uname -a, hostname, ifconfig
INFO    Fetched 1 instances for account [profile1] in [us-east-1].
INFO    Instance ID              Region          Profile         Status
INFO    i-12345					 us-east-1       profile1         Success
INFO    Execution results: 1 SUCCESS, 0 FAILED
```

#### running command(s) on multiple instances

```
> ./ssm-run -i i-12345,i-67890 -c "uname -a"

INFO    Command(s) to be executed: uname -a
INFO    Fetched 2 instances for account [profile1] in [us-east-1].
INFO    Instance ID              Region          Profile         Status
INFO    i-12345					 us-east-1       profile1         Success
INFO    i-67890					 us-east-1       profile1         Success
INFO    Execution results: 2 SUCCESS, 0 FAILED
```

#### searching for instances

By default, if no instance or filters are specified, `ssm-run` will target all instances in the current account + region.

```
> ./ssm-run -c "uname -a"

INFO    Command(s) to be executed: uname -a
INFO    Fetched 5 instances for account [profile1] in [us-east-1].
INFO    Instance ID              Region          Profile         Status
INFO    i-054151b14a3cf1234      us-east-1       profile1         Success
INFO    i-022572ac837c21234      us-east-1       profile1         Success
INFO    i-08f4076ee1fa41234      us-east-1       profile1         Success
INFO    i-0dfbe7c5db0b21234      us-east-1       profile1         Success
INFO    i-064937fab1f211234      us-east-1       profile1         Success
INFO    Execution results: 5 SUCCESS, 0 FAILED
```

#### searching for instances in multiple accounts and/or regions

```
> ./ssm-run -c "uname -a" -r us-east-1,us-west-2

INFO    Command(s) to be executed: uname -a
INFO    Fetched 5 instances for account [profile1] in [us-east-1].
INFO    Fetched 1 instances for account [profile1] in [us-west-2].
INFO    Instance ID              Region          Profile         Status
INFO    i-054151b14a3cf1234      us-east-1       profile1         Success
INFO    i-022572ac837c21234      us-east-1       profile1         Success
INFO    i-08f4076ee1fa41234      us-east-1       profile1         Success
INFO    i-0dfbe7c5db0b21234      us-east-1       profile1         Success
INFO    i-064937fab1f211234      us-east-1       profile1         Success
INFO    i-04a827989613b1234      us-west-2       profile1         Success
INFO    Execution results: 6 SUCCESS, 0 FAILED
```

#### filtering instance results

Tag-based filtering can also be applied to your search results (including if you manually specify instance names). These filters are additive, which means that each filter you provide will prune down your results to include only instances that match *all* of the provided filters.

```
> ./ssm-run -c "uname -a" -p profile1 -f app=myapp -f env=prod

INFO    Command(s) to be executed: uname -a
INFO    Fetched 6 instances for account [profile1] in [us-east-1].
INFO    Instance ID              Region          Profile         Status
INFO    i-ccdc371a               us-east-1       profile1         Success
INFO    i-9cdd364a               us-east-1       profile1         Success
INFO    i-7268418e               us-east-1       profile1         Success
INFO    i-1289eeef               us-east-1       profile1         Success
INFO    i-2f96f1d2               us-east-1       profile1         Success
INFO    i-7b6f4687               us-east-1       profile1         Success
INFO    Execution results: 6 SUCCESS, 0 FAILED
```

### usage flags

```
-all-profiles
	[USE WITH CAUTION] Parse through ~/.aws/config to target all profiles.
-c value
    --commands (shorthand)
-commands value
	Specify any number of commands to be run.
	Multiple allowed, enclosed in double quotes and delimited by semicolons (e.g. --comands "hostname; uname -a")
-dry-run
	Retrieve the list of profiles, regions, and instances your command(s) would target.
-f value
	--filter (shorthand)
-file string
	Specify the path to a shell script to use as input for the AWS-RunShellScript document.
	This can be used in combination with the --commands/-c flag, and will be run after the specified commands.
-filter value
	Filter instances based on tag value. Tags are evaluated with logical AND (instances must match all tags).
	Multiple allowed, delimited by commas (e.g. env=dev,foo=bar)
-i value
	--instances (shorthand)
-instances value
	Specify what instance IDs you want to target.
	Multiple allowed, delimited by commas (e.g. --instances i-12345,i-23456)
-limit int
	Set a limit for the number of instance results returned per profile/region combination (0 = no limit)
-log-level int
	Sets verbosity of output:
	0 = quiet, 1 = terse, 2 = standard, 3 = debug (default 2)
-p value
	--profiles (shorthand)
-profiles value
	Specify a specific profile to use with your API calls.
	Multiple allowed, delimited by commas (e.g. --profiles profile1,profile2)
-r value
	--regions (shorthand)
-regions value
	Specify a specific region to use with your API calls.
	This option will override any profile settings in your config file.
	Multiple allowed, delimited by commas (e.g. --regions us-east-1,us-west-2)

	[NOTE] Mixing --profiles and --regions will result in your command targeting every matching instance in the selected profiles and regions.
	e.g., "--profiles foo,bar,baz --regions us-east-1,us-west-2,eu-east-1" will target instances in each of the profile/region combinations:
		"foo@us-east-1, foo@us-west-2, foo@eu-east-1"
		"bar@us-east-1, bar@us-west-2, bar@eu-east-1"
		"baz@us-east-1, baz@us-west-2, baz@eu-east-1"
	Please be careful.
```
