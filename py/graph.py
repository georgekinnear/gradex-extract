#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Created on Fri Apr 24 00:52:34 2020

@author: tim
"""

import pandas as pd

df3 = pd.read_csv("../secret/third-year.csv",header=0)
df4 = pd.read_csv("../secret/fourth-year.csv",header=0)
df5 = pd.read_csv("../secret/fifth-year.csv",header=0)
dfm = pd.read_csv("../secret/msc-year.csv",header=0)

df = pd.concat([df3,df4,df5, dfm])

scanPerfect = df["ScanPerfect"]
N = len(scanPerfect)

print("Overall (3/4/5th/Msc) N=%d"%N)

tot = 0
N = len(scanPerfect)
for key,val in scanPerfect.iteritems():
    if type(val)==type(True):
        if val:
            tot = tot + 1
    elif type(val)==type("true"):
        v = val.lower()
        if v[0] =="t":
            tot = tot + 1
    else:
        N=N-1


spf = tot / N

print("Good scan %.0f%%"%(spf*100))


tot = 0
N = len(scanPerfect)

for key,val in df["HeadingPerfect"].iteritems():
    if type(val)==type(True):
        if val:
            tot = tot + 1
    elif type(val)==type("true"):
        v = val.lower()
        if v[0] =="t":
            tot = tot + 1  
    else:
        N=N-1
            

hpf = tot / N

print("Correct heading %.0f%%"%(hpf*100))


tot = 0
N = len(scanPerfect)

for key,val in df["FilenamePerfect"].iteritems():
    if type(val)==type(True):
        if val:
            tot = tot + 1
    elif type(val)==type("true"):
        v = val.lower()
        if v[0] =="t":
             tot = tot + 1    
    else:
        N = N - 1

fpf = tot / N

print("Correct filename %.0f%% "%(fpf*100))


######################

scanPerfect = df3["ScanPerfect"]

tot = 0
N = len(scanPerfect)
print("\nThird year N=%d"%N)

for key,val in scanPerfect.iteritems():
    if type(val)==type(True):
        if val:
            tot = tot + 1
    elif type(val)==type("true"):
        v = val.lower()
        if v[0] =="t":
            tot = tot + 1
    else:
        N=N-1

N = len(scanPerfect)
spf = tot / N

print("Good scan%.0f%%"%(spf*100))


tot = 0
N = len(scanPerfect)

for key,val in df3["HeadingPerfect"].iteritems():
    if type(val)==type(True):
        if val:
            tot = tot + 1
    elif type(val)==type("true"):
        v = val.lower()
        if v[0] =="t":
            tot = tot + 1  
    else:
        N=N-1
            

hpf = tot / N

print("Correct heading %.0f%%"%(hpf*100))


tot = 0
N = len(scanPerfect)

for key,val in df3["FilenamePerfect"].iteritems():
    if type(val)==type(True):
        if val:
            tot = tot + 1
    elif type(val)==type("true"):
        v = val.lower()
        if v[0] =="t":
             tot = tot + 1    
    else:
        N = N - 1

fpf = tot / N

print("Correct filename %.0f%%"%(fpf*100))


######################

scanPerfect = df4["ScanPerfect"]

tot = 0
N = len(scanPerfect)
print("\nFourth year N=%d"%N)

for key,val in scanPerfect.iteritems():
    if type(val)==type(True):
        if val:
            tot = tot + 1
    elif type(val)==type("true"):
        v = val.lower()
        if v[0] =="t":
            tot = tot + 1
    else:
        N=N-1

N = len(scanPerfect)
spf = tot / N

print("Good scan %.0f%% "%(spf*100))


tot = 0
N = len(scanPerfect)

for key,val in df4["HeadingPerfect"].iteritems():
    if type(val)==type(True):
        if val:
            tot = tot + 1
    elif type(val)==type("true"):
        v = val.lower()
        if v[0] =="t":
            tot = tot + 1  
    else:
        N=N-1
            

hpf = tot / N

print("Correct heading %.0f%%"%(hpf*100))


tot = 0
N = len(scanPerfect)

for key,val in df4["FilenamePerfect"].iteritems():
    if type(val)==type(True):
        if val:
            tot = tot + 1
    elif type(val)==type("true"):
        v = val.lower()
        if v[0] =="t":
             tot = tot + 1    
    else:
        N = N - 1

fpf = tot / N

print("Correct filename %.0f%%"%(fpf*100))

######################

scanPerfect = df5["ScanPerfect"]

tot = 0
N = len(scanPerfect)

print("\nFifth year N=%d"%N)

for key,val in scanPerfect.iteritems():
    if type(val)==type(True):
        if val:
            tot = tot + 1
    elif type(val)==type("true"):
        v = val.lower()
        if v[0] =="t":
            tot = tot + 1
    else:
        N=N-1

N = len(scanPerfect)
spf = tot / N

print("Good scan %.0f%%"%(spf*100))


tot = 0
N = len(scanPerfect)

for key,val in df5["HeadingPerfect"].iteritems():
    if type(val)==type(True):
        if val:
            tot = tot + 1
    elif type(val)==type("true"):
        v = val.lower()
        if v[0] =="t":
            tot = tot + 1  
    else:
        N=N-1
            

hpf = tot / N

print("Correct heading %.0f%%"%(hpf*100))


tot = 0
N = len(scanPerfect)

for key,val in df5["FilenamePerfect"].iteritems():
    if type(val)==type(True):
        if val:
            tot = tot + 1
    elif type(val)==type("true"):
        v = val.lower()
        if v[0] =="t":
             tot = tot + 1    
    else:
        N = N - 1

fpf = tot / N

print("Correct filename %.0f%%"%(fpf*100))


fn = df["OriginalFilename"]

#for name in fn.iteritems():
#    if name[-4:] != ".pdf":
#        #print(name[-4:])
#################################################
print("\nMsc year N=%d"%N)
tot = 0

scanPerfect = dfm["ScanPerfect"]

for key,val in scanPerfect.iteritems():
    if type(val)==type(True):
        if val:
            tot = tot + 1
    elif type(val)==type("true"):
        v = val.lower()
        if v[0] =="t":
            tot = tot + 1
    else:
        N=N-1

N = len(scanPerfect)
spf = tot / N

print("Good scan %.0f%%"%(spf*100))


tot = 0
N = len(scanPerfect)

for key,val in dfm["HeadingPerfect"].iteritems():
    if type(val)==type(True):
        if val:
            tot = tot + 1
    elif type(val)==type("true"):
        v = val.lower()
        if v[0] =="t":
            tot = tot + 1  
    else:
        N=N-1
            

hpf = tot / N

print("Correct heading %.0f%%"%(hpf*100))


tot = 0
N = len(scanPerfect)

for key,val in dfm["FilenamePerfect"].iteritems():
    if type(val)==type(True):
        if val:
            tot = tot + 1
    elif type(val)==type("true"):
        v = val.lower()
        if v[0] =="t":
             tot = tot + 1    
    else:
        N = N - 1

fpf = tot / N

print("Correct filename %.0f%%"%(fpf*100))
