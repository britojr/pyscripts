#!/usr/bin/env Rscript
# ------------------------------------------------------------------------------
# example of option parsing
# ------------------------------------------------------------------------------
library("optparse")
 
opt_list = list(
  make_option(c("-f", "--file"), type="character", default=NULL, 
    help="dataset file name"),
  make_option(c("-o", "--out"), type="character", default="out.txt", 
    help="output file name [default= %default]")
); 
 
opt_parser = OptionParser(option_list=opt_list);
opt = parse_args(opt_parser);

# required arguments
if (is.null(opt$file)){
  print_help(opt_parser)
  stop("At least one argument must be supplied (input file)", call.=FALSE)
}

cat(sprintf("parm %s: %s\n", "-f", opt$f))
cat(sprintf("parm %s: %s\n", "-o", opt$o))
