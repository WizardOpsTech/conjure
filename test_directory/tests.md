# Tests

Manual tests to run for now that give me the warm fuzzy. Automate tests later. 

## Conjure Template
- conjure template with output and no flags = interactive mode
- conjure template without output = exit 1 prompt for an output
- conjure template with -f flag only = generate artifact
- conjure template with --var flags only
    - bad input test for non key=value pair
    - only 1 --var provided when multiple required 
    - works when all --vars provided
- conjure template with -f and --var flags combined - applys --var vars over -f vars and combines them.

## Conjure Bundle
- conjure bundle with output and no flags = interactive mode
    - additionally prompt for additional template level overrides
- conjure bundle without output = exit 1 prompt for output
- conjure bundle with -f flag only = generate artifact
- conjure bundle with --var flags only
    - bad input test for non key=value pair
    - only 1 --var provided when multiple required
    - works when all required --vars provided
- conjure bundle with -f and --var flags combined = merges the 2 
