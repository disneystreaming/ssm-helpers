# ssm run

Run a command on a list of instances or based on instance tags.
Increase `--verbose` to see individual instances or command output.
Run multiple commands in the same shell context (e.g. `-c 'uptime;hostname'`).

![](../../img/ssm-run.gif)

## about

`ssm-run` is a tool for executing commands on SSM-managed EC2 instances without having to connect to them directly.

### basic usage

By default, `ssm run` will attempt to search/connect for instances in your default AWS profile (specified in your ~/.aws/config, or via the `AWS_PROFILE` environment variable if set), and the default region linked to that profile (or the `AWS_REGION` environment variable, if set).

The order of preference for profile/region settings is set as follows:

1) explicitly specified by the --profiles and/or --regions flags
2) explicitly set using the `AWS_PROFILE` and/or `AWS_REGION` environment variables
3) falling back to the default profile and/or region as specified in your ~/.aws/config file.

To use a specific profile and/or region, use the `-p (--profiles)` or `-r (--regions)` flags.

e.g. `-p account1,account2 -r us-east-1,us-west-2`

Take careful note that if multiple regions *and* profiles are specified, `ssm run` will attempt to run your commands across all the possible permutations.

For example, using the flags `--profiles foo,bar,baz --regions us-east-1,us-west-2,eu-east-1` will target instances in each of the profile/region combinations:

	foo@us-east-1, foo@us-west-2, foo@eu-east-1
	bar@us-east-1, bar@us-west-2, bar@eu-east-1
	baz@us-east-1, baz@us-west-2, baz@eu-east-1

As such, please be careful when doing so and when using the `--all-profiles` flag.

If a region is not specified, you will receive the following warning, but execution will continue:

```
WARNING	No AWS region setting found. Will default to the region linked to any profile in use.
```

#### running command(s) on a single instance

```
> ssm run -p 'profile1' -i i-12345 -c 'uname'
INFO    Command(s) to be executed:
uname             
INFO    Started invocation f73f2225-8fb2-4e63-ba63-6e2af54b8659 for profile1 in us-east-1 
INFO    Instance ID              Region          Profile         Status 
INFO    i-12345                  us-east-1       profile1        Success 
INFO    Linux                                        
INFO    Execution results: 1 SUCCESS, 0 FAILED
```

Running multiple commands is easy; just separate your commands with semicolons (;). These commands will be executed in series, in the same console context (e.g., variables you set can be used between commands).

```
> ssm run -p 'profile1' -i i-12345 -c 'uname;uname;uname'
INFO    Command(s) to be executed:
uname
uname
uname 
INFO    Started invocation 73ee2f4c-fd54-4505-8ef7-1bb2baecd64f for profile1 in us-east-1 
INFO    Instance ID              Region          Profile         Status 
INFO    i-12345                  us-east-1       profile1        Success 
INFO    Linux
Linux
Linux                            
INFO    Execution results: 1 SUCCESS, 0 FAILED 
```

#### running command(s) on multiple instances

```
> ssm run -p 'profile1' -i 'i-0879fe217fe11ad86,i-062aede5be23eef7a' -c 'uname -sm'
INFO    Command(s) to be executed:
uname > /dev/null 2>&1 
INFO    Started invocation cda5592a-a099-4117-8863-32a88909eae6 for profile1 in us-east-1 
INFO    Instance ID              Region          Profile         Status 
INFO    i-12345                  us-east-1       profile1        Success
Linux
INFO    i-23456                  us-east-1       profile1        Success 
Linux
INFO    Execution results: 2 SUCCESS, 0 FAILED
```

#### searching for instances

By default, if no instance or filters are specified, `ssm-run` will target all instances in the current account + region.

```
> ssm run -p 'profile1' -c 'uname > /dev/null 2>&1'
INFO    Command(s) to be executed:
uname > /dev/null 2>&1 
INFO    Started invocation a0fc81ce-a256-4b19-803f-8b24e453172d for profile1 in us-east-1 
INFO    Instance ID              Region          Profile         Status 
INFO    i-12345                  us-east-1       profile1        Success 
INFO    i-23456                  us-east-1       profile1        Success 
INFO    i-34567                  us-east-1       profile1        Success 
INFO    Execution results: 3 SUCCESS, 0 FAILED
```

#### searching for instances in multiple accounts and/or regions

```
> ssm run -p 'profile1' -c 'uname > /dev/null 2>&1' --region 'us-east-1,us-west-2'
INFO    Command(s) to be executed:
uname > /dev/null 2>&1
INFO    Started invocation b94eafc1-c9ab-4f9b-848e-c4e16beecee2 for profile1 in us-east-1
INFO    Started invocation 83a0a57b-1127-4ff8-9fb9-136f040a05fd for profile1 in us-west-2
INFO    Instance ID              Region          Profile         Status
INFO    i-12345                  us-east-1       profile1        Success
INFO    i-23456                  us-west-2       profile1        Success
INFO    Execution results: 2 SUCCESS, 0 FAILED
```

#### filtering instance results

Tag-based filtering can also be applied to your search results (including if you manually specify instance names). These filters are additive, which means that each filter you provide will prune down your results to include only instances that match *all* of the provided filters.

```
> ssm run -p 'profile1' -f 'app=myapp,env=prod' -c 'uname > /dev/null 2>&1' --region us-east-1
INFO    Command(s) to be executed:
uname > /dev/null 2>&1 
INFO    Started invocation 1a781eaf-a6fc-4cf0-8875-5ecaace29e4f for profile1 in us-east-1 
INFO    Instance ID              Region          Profile         Status 
INFO    i-12345                  us-east-1       profile1        Success 
INFO    Execution results: 1 SUCCESS, 0 FAILED
```

### usage flags

```
--all-profiles
	[USE WITH CAUTION] Parse through ~/.aws/config to target all profiles.
-c, --command string
	Specify any number of commands to be run.
	Multiple allowed, enclosed in double quotes and delimited by semicolons (e.g. --comands "hostname; uname -a")
--dry-run
	Retrieve the list of profiles, regions, and instances your command(s) would target
--file string
	Specify the path to a shell script to use as input for the AWS-RunShellScript document.
	his can be used in combination with the --commands/-c flag, and will be run after the specified commands.
-f, --filter strings
	Filter instances based on tag value. Tags are evaluated with logical AND (instances must match all tags).
	Multiple allowed, delimited by commas (e.g. env=dev,foo=bar)
-h, --help 
	help for run
-i, --instance strings
	Specify what instance IDs you want to target.
	Multiple allowed, delimited by commas (e.g. --instance i-12345,i-23456)
--max-concurrency string
	Max targets to run the command in parallel. Both numbers, such as 50, and percentages, such as 50%, are allowed (default "50")
--max-errors string
	Max errors allowed before running on additional targets. Both numbers, such as 10, and percentages, such as 10%, are allowed (default "0")
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
```
