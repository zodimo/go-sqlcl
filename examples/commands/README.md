# SQLcl Command Wrappers Example

This example demonstrates how to use the SQLcl-specific command wrappers provided by the go-sqlcl package. These wrappers make it easier to work with SQLcl-specific functionality in a more idiomatic Go way.

## Overview

The example shows how to:

1. Create and configure a SQLcl client
2. Connect to a database (or run in demo mode)
3. Use various SQLcl-specific commands:
   - SET commands (PAGESIZE, LINESIZE, TIMING, etc.)
   - SHOW commands
   - DESCRIBE command to get table structure
   - HELP command
   - HISTORY command to view command history

## Running the Example

### Prerequisites

- Go 1.16 or later
- Oracle SQLcl installed and accessible on your PATH, or set a custom path in the example code
- Oracle database (optional - the example can run in demo mode without a database)

### With a Database Connection

To run the example with a real database connection, set the following environment variables:

```bash
export ORACLE_USERNAME=your_username
export ORACLE_PASSWORD=your_password
export ORACLE_CONNECT_STRING=host:port/service_name
```

Then run the example:

```bash
go run main.go
```

### Demo Mode

If you don't have a database available, you can just run the example without setting environment variables:

```bash
go run main.go
```

The example will run in demo mode and show how the commands would be used.

## Available Command Wrappers

The go-sqlcl package provides the following command wrappers:

### SET Commands

- `SetPageSize(ctx, size)`: Sets the number of rows displayed per page
- `SetLineSize(ctx, size)`: Sets the line width
- `SetFeedback(ctx, enabled)`: Enables or disables command feedback
- `SetTiming(ctx, enabled)`: Enables or disables timing information
- `SetSQLTerminator(ctx, terminator)`: Sets the SQL terminator character
- `SetHeading(ctx, enabled)`: Enables or disables column headings
- `SetTNSAdmin(ctx, directory)`: Sets the TNS_ADMIN directory
- `SetCommand(ctx, option, value)`: Generic method to set any option

### SHOW Command

- `ShowCommand(ctx, option)`: Shows the current value of a specified option

### DESCRIBE Command

- `DescribeObject(ctx, objectName)`: Describes a database object (table, view, etc.)
  - Returns a structured `DescribeResult` with all column information

### HELP Command

- `GetHelp(ctx, topic)`: Gets help information on a specific topic

### HISTORY Command

- `GetHistory(ctx)`: Gets the command history as a slice of strings

## Customizing the Example

You can modify the example to:

- Change the SQLcl path if your SQLcl installation is in a non-standard location
- Adjust the timeout values to suit your environment
- Change the table name in the DESCRIBE example to match a table in your database
- Add more examples using other command wrappers provided by the package 