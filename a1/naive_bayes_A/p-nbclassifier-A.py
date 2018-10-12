#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
CSC564 (Fall 2018) - Assignment 1 - Problem 6
Parallel Multinominal Naive Bayes classifier for Text Classification.
=========================================
Implementation A: Multiprocessing

Created on Fri Oct 02 2018

@author: Spencer Rose

References:
Code heavily adapted from classifier written by Syed Sadat Nazrul (retrieved Sept 20, 2018)
https://towardsdatascience.com/multinomial-naive-bayes-classifier-for-text-analysis-python-8dd6825ece67
"""
import pandas as pd
import numpy as np
from time import time
from multiprocessing import Process, Queue
import sys
  
    
#===================================== 
def main():
  
    ts = time()

    # Get training data: create Pandas dataframe
    train_data = open('data/20news-bydate/matlab/train.data', 'r', newline='')
    df = pd.read_csv(train_data, delimiter=' ', names=['docIdx', 'wordIdx', 'count'])
    
    # Get training labels
    train_label = open('data/20news-bydate/matlab/train.label', 'r', newline='').readlines()
    
    # Extract vocabulary
    vocab_df = extract_vocab('data/20news-bydate/vocabulary.txt')
            
    # Build training model  
    print("Building training model...")
    df, pi, Pr_dict = nb_multi_train(df, train_label, vocab_df)
        
    print("Applying model to training data...")
    prediction = nb_multi_apply(df, pi, Pr_dict)
   
    #Get list of labels
    train_label = pd.read_csv('data/20news-bydate/matlab/train.label', names=['t'])
    train_label= train_label['t'].tolist()
    calc_accuracy(prediction, train_label)
    
    print('Total running time (s):', time() - ts)        

# =====================================
# Multinominal Naive Bayes training algorithm
def nb_multi_train (df, train_label, vocab_df):
    ts = time()
    
    # initialize prior probabilities dictionary (fraction of total docs)
    prior = {i:0 for i in range(1,21)}
    n_docs = len(train_label)
    print("Total number of training documents = %d\n" % n_docs)
        
    for key in prior:
        n_docs_in_class = len([a for a in train_label if int(a) == key])
        prior[key] = n_docs_in_class/n_docs
        # print("Fraction of training documents in class {:d} = {:5.4f} ({:d})\n".format(key, prior[key], n_docs_in_class))
        
    #Training labels
    label = []
    for line in train_label:
        label.append(int(line.split()[0]))
    
    #Increase label length to match docIdx
    docIdx = df['docIdx'].values
    i = 0
    new_label = []
    for index in range(len(docIdx)-1):
        new_label.append(label[i])
        if docIdx[index] != docIdx[index+1]:
            i += 1
    new_label.append(label[i]) #for-loop ignores last value
    
    #Add label column
    df['classIdx'] = new_label

    #Alpha value for Laplace smoothing
    a = 0.01
    v_count = len(vocab_df)
    
    #Calculate probability of each word based on class
    pb_ij = df.groupby(['classIdx','wordIdx'])
    pb_j = df.groupby(['classIdx'])
    Pr =  (pb_ij['count'].sum() + a) / (pb_j['count'].sum() + v_count)    

    #Unstack series: make df have new level of column labels 
    Pr = Pr.unstack()
    
    #Replace NaN or columns with 0 as word count with a/(count+|V|+1)
    for c in range(1,21):
        Pr.loc[c,:] = Pr.loc[c,:].fillna(a/(pb_j['count'].sum()[c] + v_count))
    
    #Convert to dictionary for greater speed
    Pr_dict = Pr.to_dict()
    
    #Index of all words
    tot_list = set(vocab_df['index'])
    
    #Index of good words
    vocab_df = vocab_df[~vocab_df['word'].isin(stop_words)]
    good_list = vocab_df['index'].tolist()
    good_list = set(good_list)
    
    #Index of stop words
    bad_list = tot_list - good_list
    
    q = Queue() # input queue
    r = Queue() # results
        
    for j in range(1, 21):
        Process(target=calc_pi, args=(q, r,)).start()        
    
    #BOTTLENECK 1: Set all stop words to 0
    for bad in bad_list:
        for j in range(1, 21):
            q.put((bad, j, 21, a, pb_j, v_count))
    
    while not q.empty():
         result = r.get()
         Pr_dict[result[0]][result[1]] = result[2]
         
    print('Training running time (s):', time() - ts)
    return df, prior, Pr_dict

# =======================================
def nb_multi_apply(df, pi, Pr_dict):
    ts = time()
    '''
    Multinomial Naive Bayes classifier
    :param df [Pandas Dataframe]: Dataframe of data
    :return prediction [list]: Predicted class ID
    '''
    #Using dictionaries for greater speed
    df_dict = df.to_dict()
    new_dict = {}
    prediction = []
    
    
    #new_dict = {docIdx : {wordIdx: count},....}
    for idx in range(len(df_dict['docIdx'])):
        docIdx = df_dict['docIdx'][idx]
        wordIdx = df_dict['wordIdx'][idx]
        count = df_dict['count'][idx]
        try: 
            new_dict[docIdx][wordIdx] = count 
        except:
            new_dict[df_dict['docIdx'][idx]] = {}
            new_dict[docIdx][wordIdx] = count

    #Calculating the scores for each doc
    for docIdx in range(1, len(new_dict)+1):
        score_dict = {}
        #Creating a probability row for each class
        for classIdx in range(1,21):
            score_dict[classIdx] = 1
            #For each word:
            for wordIdx in new_dict[docIdx]:
                try:
                    probability = Pr_dict[wordIdx][classIdx]        
                    power = new_dict[docIdx][wordIdx]               
                    score_dict[classIdx]+=power*np.log(probability)                     
                except:
                    #Missing V will have 0*log(a/16689) = 0
                    score_dict[classIdx] += 0      
            #Multiply final with pi         
            score_dict[classIdx] +=  np.log(pi[classIdx])                          

        #Get class with max probabilty for the given docIdx 
        max_score = max(score_dict, key=score_dict.get)
        prediction.append(max_score)
    '''
    q = Queue() # input queue
    r = Queue() # results
    # Generate processes
    for j in range(1, 21):
        Process(target=calc_score, args=(q, r,)).start()
        
    BOTTLENECK 2: Calculating the scores for each doc
    #Calculating the scores for each doc
    for docIdx in range(1, len(new_dict)+1):
        score_dict = {}

        #Creating a probability row for each class
        for classIdx in range(1,21):
            q.put((new_dict, Pr_dict, docIdx, classIdx, pi))
    
        while not r.empty() or not q.empty():
            result = r.get()
            score_dict[result[1]] = result[0]
                     
        #Get class with max probabilty for the given docIdx 
        max_score = max(score_dict, key=score_dict.get)
        prediction.append(max_score)
    '''
        
    print('Testing running time (s):', time() - ts)
        
    return prediction

# =======================================
# Returns extracted vocabulary from prepared list    
def extract_vocab(fname):
    vocab_data = open(fname)
    vocab_df = pd.read_csv(vocab_data, names = ['word'])
    vocab_df = vocab_df.reset_index()
    vocab_df['index'] = vocab_df['index'].apply(lambda x: x+1) 
    return vocab_df

# =======================================
# Calculates accuracy of predictions based on given class labels
def calc_accuracy (predicted_label, test_label):
    n = len(test_label)
    val = 0
    for i,j in zip(predicted_label, test_label):
        if i == j:
            val +=1
        else:
            pass   
    print("Prediction Accuracy:\t\t",val/n * 100, "%")


# =======================================
# Bottleneck 1: Stop word removal
# =======================================
def calc_pi(q, r):
    while 1:
        if not q.empty():
            (bad, j, cls, a, pb_j, v_count) = q.get()
            value = a/(pb_j['count'].sum()[j] + v_count)
            r.put((j, bad, value))
            sys.exit

     
# =======================================    
# Bottleneck 2: Score calculation
def calc_score(q, r):
    while 1:
        if not q.empty():
            (new_dict, Pr_dict, docIdx, classIdx, pi) = q.get()
            value = 1
            for wordIdx in new_dict[docIdx]:
                try:
                    probability = Pr_dict[wordIdx][classIdx]        
                    power = new_dict[docIdx][wordIdx]               
                    value += power*np.log(probability)                     
                except:
                    #Missing V will have 0*log(a/16689) = 0
                    value += 0
            #Multiply final with pi         
            value +=  np.log(pi[classIdx])
            r.put((value, classIdx))
    
#Common stop words from online
stop_words = [
"a", "about", "above", "across", "after", "afterwards", 
"again", "all", "almost", "alone", "along", "already", "also",    
"although", "always", "am", "among", "amongst", "amoungst", "amount", "an", "and", "another", "any", "anyhow", "anyone", "anything", "anyway", "anywhere", "are", "as", "at", "be", "became", "because", "become","becomes", "becoming", "been", "before", "behind", "being", "beside", "besides", "between", "beyond", "both", "but", "by","can", "cannot", "cant", "could", "couldnt", "de", "describe", "do", "done", "each", "eg", "either", "else", "enough", "etc", "even", "ever", "every", "everyone", "everything", "everywhere", "except", "few", "find","for","found", "four", "from", "further", "get", "give", "go", "had", "has", "hasnt", "have", "he", "hence", "her", "here", "hereafter", "hereby", "herein", "hereupon", "hers", "herself", "him", "himself", "his", "how", "however", "i", "ie", "if", "in", "indeed", "is", "it", "its", "itself", "keep", "least", "less", "ltd", "made", "many", "may", "me", "meanwhile", "might", "mine", "more", "moreover", "most", "mostly", "much", "must", "my", "myself", "name", "namely", "neither", "never", "nevertheless", "next","no", "nobody", "none", "noone", "nor", "not", "nothing", "now", "nowhere", "of", "off", "often", "on", "once", "one", "only", "onto", "or", "other", "others", "otherwise", "our", "ours", "ourselves", "out", "over", "own", "part","perhaps", "please", "put", "rather", "re", "same", "see", "seem", "seemed", "seeming", "seems", "she", "should","since", "sincere","so", "some", "somehow", "someone", "something", "sometime", "sometimes", "somewhere", "still", "such", "take","than", "that", "the", "their", "them", "themselves", "then", "thence", "there", "thereafter", "thereby", "therefore", "therein", "thereupon", "these", "they",
"this", "those", "though", "through", "throughout",
"thru", "thus", "to", "together", "too", "toward", "towards",
"under", "until", "up", "upon", "us",
"very", "was", "we", "well", "were", "what", "whatever", "when",
"whence", "whenever", "where", "whereafter", "whereas", "whereby",
"wherein", "whereupon", "wherever", "whether", "which", "while", 
"who", "whoever", "whom", "whose", "why", "will", "with",
"within", "without", "would", "yet", "you", "your", "yours", "yourself", "yourselves"
]

# Encapsulate in main()
if __name__ == "__main__":
    main()