#!/usr/bin/env Rscript
# ------------------------------------------------------------------------------
# randomly generate a number of queries for root values
# ------------------------------------------------------------------------------
library(bnlearn)
library(optparse)

# define arguments
opt_list = list(
  make_option(c("-m", "--model"), type="character", default=NULL,
    help="model file name"),
  make_option(c("-q", "--query"), type="character", default=NULL,
    help="query file name"),
  make_option(c("-n", "--num"), type="integer", default=1,
    help="number of queries to sample [default= %default]")
)
opt_parser = OptionParser(option_list=opt_list)
opt = parse_args(opt_parser)

# test required arguments
if (is.null(opt$model) || is.null(opt$query)){
  print_help(opt_parser)
  stop("missing arguments", call.=FALSE)
}

# read bnet
bn = read.bif(opt$model, debug = FALSE)
# get bn roots
rs = root.nodes(bn)

print(roots)
sample(rs, opt$num, replace=TRUE)
