#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Created on Fri Apr 24 19:48:16 2020

@author: tim
"""

import pandas as pd

dfin = pd.read_csv("../secret/ingest-report-1587778748523537564.csv")

dfchk = pd.read_csv("../secret/msc-year-updated-v2.csv")


#dfin.columns
#
#Index(['FirstName', 'LastName', 'Matriculation', 'Assignment', 'DateSubmitted',
#       'SubmissionField', 'Comments', 'OriginalFilename', 'Filename',
#       'ExamNumber', 'MatriculationError', 'ExamNumberError', 'FiletypeError',
#       'FilenameError', 'NumberOfPages', 'FilesizeMB', 'NumberOfFiles'],
#      dtype='object')
#dfchk.columns
# 
#Index(['ScanPerfect', 'ScanRotated', 'ScanContrast', 'ScanFaint',
#       'ScanIncomplete', 'ScanBroken', 'ScanComment1', 'ScanComment2',
#       'HeadingPerfect', 'HeadingVerbose', 'HeadingNoLine',
#       'HeadingNoQuestion', 'HeadingNoExamNumber', 'HeadingAnonymityBroken',
#       'HeadingComment1', 'HeadingComment2', 'FilenamePerfect',
#       'FilenameVerbose', 'FilenameNoCourse', 'FilenameNoID', 'InputFile',
#       'BatchFile', 'BatchPage', 'FirstName', 'LastName', 'Matriculation',
#       'Assignment', 'DateSubmitted', 'SubmissionField', 'Comments',
#       'OriginalFilename', 'Filename', 'ExamNumber', 'MatriculationError',
#       'ExamNumberError', 'FiletypeError', 'FilenameError', 'NumberOfPages',
#       'FilesizeMB', 'NumberOfFiles', 'Submission'],
#      dtype='object')
   

notFoundMatric = 0
missingMatrics = []

indices = []


for i, imatric in dfin["Matriculation"].iteritems():
    match = False
    for j, cmatric in dfchk["Matriculation"].iteritems():
        if imatric == cmatric:
            match = True
    if not match:
        notFoundMatric = notFoundMatric + 1
        missingMatrics.append(imatric)
        indices.append(i)
        #print("Didn't find %s\n"%(imatric))
        


for i in indices:
    f = dfin.iloc[i,:]["Filename"].split('.')[0]
    of = dfin.iloc[i,:]["OriginalFilename"].split('.')[0]
    match = False
    for j, filename in dfchk["OriginalFilename"].iteritems():
        if type(filename)==str:
            fc = filename.split('.')[0]
            if (f == fc) or (of == fc):
                #print("%s matched to %s\n"%(mm,fc))
                match = True
    if match == False:
        print("Didn't find files for %s, tried\n%s\n%s\n"%(dfin["Matriculation"][i],of,f))
            
    

           