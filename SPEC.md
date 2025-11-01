# ifdb-dl

Search and download story files from IFDB.

# Flags

We should support search by genre, format, etc, etc. Other than the obvious...

* -h/--help
  * request help
* -o/-output
  * directory to put file
  
If there is only one match in the search results we should just go ahead and download that one.
Otherwise prompt the user for which one to download. If there are an insane amount of results tell user
that there are too many and have them make a more specific search.

## -i/--interactive flag
Makes the program walk the user through the process of searching for and downloading the game.
