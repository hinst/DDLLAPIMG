# Delphi DLL API macro-generator

In processed file:

	{$region function headers}
	// put function headers here, like
	function do_something(argument: LongInt): LongInt; stdcall;
	{$endRegion function headers}
	
	{$region function loader template}
	// put function loader template here
	{$endRegion function loader template}

	{$region deferred functions}
	// here generated deferred function loaders will be inserted
	{$endRegion deferred functions}


