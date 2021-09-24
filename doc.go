// Copyright (c) 2021 CoderMe.com
// Permission to use, copy, modify, and distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
// WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
// ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
// WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
// ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
// OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.

/*
	Package oneid provides utility functions for generating unique numeric IDs concurrently and
        safely distributable across multiple servers.



	Generating IDs:

	IDs generated are partially time-sortable, guranteed to be unique only if the program is running by one
        process and on a single server. On cluster of servers or multiple processes per server, further configuration
        is needed.

        Int64(serverID, processID, Config) and Uint64(serverID, processID, config) can be used to generate IDS with
        fixed serverID and processID on simple setups.

        For more two servers or more, there are os nviroment depenedant which lookup up serverID is environment
        SERVER_ID and processID in enviroment variable PROCESS_ID


        Arguments:

	Both ServerID and ProcessID accept positive numbers olny as valid values, besides processID accepts zero
        as valid value too which indicates the function to use the current dynamic system process id (pid).



	Default Configurations:

        The default configration supports upto 1024 servers and upto 32 processes per each one.

        To support more than 1024 servers, or more than 32 processes, consider customizing  processBits and serverBits
        by using NewInt64Config() and NewUint64Config() for Int64(), EnvInt64() and its uint equivelent Uint64()
        EnvUint64() functions respectively.

        Configurations are thread-safe and should be reused across multiple goroutines.



	Limitations:

        Duplicate id maybe generated on heavly-loaded setups, to minimize the possibility of generating
        duplicates:


          * One server with multiple processes:
             = Consider using processes manager like systemd, in order to decrease the os dynamic proces sid
             aka (pid) gap, large gaps between pid increase the likelihood of exhasting processbits.

             = A better solution would be use static processID as environment variable available for a single
             process each. Then use EnvInt64 and EnvUint64.


          * Multiple servers with multiple processe:
             = Avoid out of range server ids.



*/
package oneid

