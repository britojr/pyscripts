#!/usr/bin/env Rscript
# ------------------------------------------------------------------------------
# generate data samples of a bayesian network
# ------------------------------------------------------------------------------
library(bnlearn)

# read arguments
args = commandArgs(trailingOnly=TRUE)
if (length(args) < 2) {
  print("USAGE:")
  print("Rscript sample.R file.bif numSampes")
  stop("missing arguments", call.=FALSE)
}
fname = args[1]
nSamp = as.integer(args[2])

# create output file name
dsfile = paste(tools::file_path_sans_ext(fname), ".csv", sep="")

# read bnet
bn = read.bif(fname, debug = FALSE)

# bnet sample
df = rbn(bn, nSamp, debug = FALSE)

# convert dataframe values to binary
write.csv(df, dsfile, quote=FALSE, row.names=FALSE)
df = read.csv(dsfile, header=TRUE, stringsAsFactors=FALSE)
df[df=="no"]<-"0"
df[df=="yes"]<-"1"
# save in csv format
write.csv(df, dsfile, quote=FALSE, row.names=FALSE)

