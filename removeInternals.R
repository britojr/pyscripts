#!/usr/bin/env Rscript
# ------------------------------------------------------------------------------
# remove internal variables from a generated dataset
# ------------------------------------------------------------------------------
library(bnlearn)

# read arguments
args = commandArgs(trailingOnly=TRUE)
if (length(args) < 3) {
  print("USAGE:")
  print("Rscript sample.R file.bif file.csv outfile.csv")
  stop("missing arguments", call.=FALSE)
}
bnfile = args[1]
csvfile = args[2]
outfile = args[3]

# read bnet and dataset
bn = read.bif(bnfile, debug = FALSE)
df = read.csv(csvfile, header=TRUE, stringsAsFactors=FALSE)

# keep only the outer nodes (roots and leafs)
outer = c(leaf.nodes(bn), root.nodes(bn))
df = df[,(names(df) %in% outer )]

# save new csv
write.csv(df, outfile, quote=FALSE, row.names=FALSE)

