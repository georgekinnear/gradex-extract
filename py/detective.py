#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Created on Fri Apr 24 19:48:16 2020

@author: tim
"""

import pandas as pd

df = pd.read_csv("../secret/beforeafter.csv")

df1 = pd.read_csv("../secret/ingest-report-1587611163557432472.csv")
df2 = pd.read_csv("../secret/ingest-report-1587659335206699389.csv")

ingest = pd.concat([df2,df2])

df.columns
ingest.columns

trimlen = len("Practice Exam Drop Box_s2032738_attempt_2020-04-22-04-19-43_")

for i, before in df["Before"].iteritems():
    expectedName = before[trimlen:]
    match = False
    for j, after in df["After"].iteritems():
        if expectedName == after:
            match = True
    if not match:
        print("Didn't find %s as %s\n"%(before,expectedName))
        
        
Nb = 0        
for i, file in df["Before"].iteritems():
    if not pd.isna(file):
        Nb = Nb + 1

    
    
Na = 0        
for i, file in df["After"].iteritems():
    if not pd.isna(file):
        Na = Na + 1

    
print("N(before)=%d\nN(after)=%d\n"%(Nb,Na))    