#!/usr/bin/python3
# ----------------------------------------------------------------------
# random split of a dataset in train and test
# ----------------------------------------------------------------------
import sys
import os
import pandas as pd
import numpy as np
from sklearn.model_selection import train_test_split

def extendName(fname, add):
   base, ext = os.path.splitext(fname)
   return base + add + ext

def main():
   # check input
   if len(sys.argv) < 2:
      print('usage:')
      print('python3 datasplit.py data.csv test_size')
      sys.exit(0)
   else:
      print('arguments: ', sys.argv)

   # read arguments
   if len(sys.argv) < 3:
      tsize = 0.2
   else:
      tsize = float(sys.argv[2])
   
   # read data
   fname = sys.argv[1]
   df = pd.read_csv(fname)
   print(fname, ' shape: ', df.shape)

   # split data
   train, test = train_test_split(df, test_size=tsize)

   # save train and test data
   fname_train, fname_test = extendName(fname, '_train'), extendName(fname, '_test')
   train.to_csv(fname_train, index=False)
   test.to_csv(fname_test, index=False)
   print(fname_train, ' shape: ', train.shape)
   print(fname_test, ' shape: ', test.shape)


main()
