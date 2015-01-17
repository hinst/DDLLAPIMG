# Delphi DLL API macro-generator

In processed file:

```delphi

	{$region function headers}
	// put function headers here, like
	function do_something(argument, brgument: LongInt): LongInt; stdcall;
	{$endRegion function headers}
	
	{$region function loader template}
	// put function loader template here; 
	// these are substituted:
	$routineKind$ -> function or procedure
	$routineName$
	$routineTail$ -> "(argument, brgument: LongInt): LongInt; stdcall;"
	$routineArguments$ -> argument, brgument (argument names separated by comma with no types)
	{$endRegion function loader template}
	
	{$region deferred functions}
	// here generated deferred function loaders will be inserted
	{$endRegion deferred functions}
	
```

Sample function loader template:

```delphi

	type T$routineName$ = $routineKind$ $routineTail$ {routineTail with ; }
	var $routineName$_routine: T$routineName$;
	$routineKind$ $routineName$ $routineTail$
	begin
		if GetRoutine(@$routineName$_routine, '$routineName$') then
			$resultAssignmentPrefixIfFunction$$routineName$_routine($routineArguments$);
	end;
	
```
