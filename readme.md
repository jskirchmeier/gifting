# Generate a Gift List

Many families (or other groups) draw names out of a hat for assigning someone to buy a gift for.  
This creates a (mostly) random selection each year.  There are some problems with that

1. You can draw the same name as you did last year
1. You may have to throw a name back in since you don't want to draw your spouse (since you have to buy them a gift anyway)
1. Not everyone may be available for the drawing

In order to crate the optimum gift giving assignment list a few simple rules are needed:

1.  A person cannot receive more than one gift
1.  A person cannot give more than one gift
1.  You cannot be assigned someone in your 'family' (spouse, children), but only those outside
    the family (siblings, nephews, ...)
1.  The number of repeat assignments should be kept to a minimum
    *   More recent repeats count more that old repeats
    *   Someone with fewer past repeats should be given a repeat over someone with more past repeats
1.  Reciprocal assignments are not allowed (if John gets Jane this year then Jane cannot get John)


**Simple right!**


### A little history
In the beginning the pool was quite small.  Lets look at an example of  3 couples, each person has four possible others 
that they can gift to. There are 24 legal parings and 4,096 possible (though not all legal, from the rules above) solutions.

```
Families                  :                                         3
Population                :                                         6
Possible Pairs            :                                        24
Possible Solutions        :                                     4,096
```

This can be done manually and was kept a spreadsheet, updating it each year and manually figuring out with the new 
parings would be. As children grow they joined the 'adults' pool (in our family it was the year after graduation from high school,
so you not only had to become an adult but stopped getting presents from all of your aunts and uncles and now just got
one and had to buy one also).  The kids get married and the spouses join the pool.   This manual process became quite a 
burden and sub-optimal assignments were made (these bad early decisions still haunt the current calculations as all 
history is needed to make the best assignments).

The numbers for a pool of 14 persons in family units of 3, 4, 5, and 2 give us :

```
Families                  :                                         4
Population                :                                        14
Possible Pairs            :                                       142
Possible Solutions        :                       113,175,675,360,000
```

Soon it became apparent we needed to automate!  So a simple brute force Java application was written.  
It determined which of the possible solutions were legal and then scored them.  The one with the best score won!  
Not efficient but what the heck it only had to be run once a year and so what if it took several minutes to do it's thing.

The calculations were taking hours and then when it ran overnight without completion I killed it and took the best 
solution found so far.  For the following year a new Java program was used that did some smart graph eliminations and 
reduced that time to seconds.  This has been used for a number years.

The pool continued to grow and the time to calculate the solution also grew (and not linearly).  Our pool is now   
families of size 5, 6, 9, and 2 for 22 total in the pool.

```
Families                  :                                         4
Population                :                                        22
Possible Pairs            :                                       338
Possible Solutions        :       101,044,962,002,466,589,009,510,400
```

That program has now been rewritten in GO and here we are!


### A daunting task
 
With the example above (22 folks across 4 family groups) a **very** rough calculation gives us many thousands of years
to brute force the calculation on my desktop system.  Even multi-threading it on my 4 core in a perfect world would 
still leave us waiting a long time for our presents.

> As a side note, this problem did not lend itself to parallel processing.  Each unit of work was quite small
> and the overhead of synchronization actually caused it to take longer.  I tried many different ways, tightly 
> coupled and the overhead was too large.  Loose coupling hindered the effectiveness of the reduction algorithm and 
> resulted in more work being done.  The end result was a well constructed (if I say so myself) single threaded 
> program worked very well indeed (see the numbers below).


This iteration of the program is able to calculate the optimum solution in 59ms (with the example 22 person pool above).
WOW!  That's pretty good, how'd you do that?  I'm glad you asked!


###### Get the data

We start with an XML file that defines the pool, with family groupings and the history of any past giving.
It looks like :

```
<DataStore>
  <family>
    <person name="Adam" >
      <history recipient="Barry" year="2010"></history>
      <history recipient="Cindy" year="2011"></history>
    </person>
    <person name="Ann" >
      <history recipient="Charles" year="2010"></history>
      <history recipient="David" year="2011"></history>
    </person>
  </family>
  <family>
    <person name="Barry">
      <history recipient="Ann" year="2010"></history>
      <history recipient="Adam" year="2011"></history>
    </person>
    <person name="Betty">
      <history recipient="Adam" year="2010"></history>
      <history recipient="Cindy" year="2011"></history>
    </person>
  </family>
  ...
</DataStore>
```


###### Build our data structure

Figure out whom each person can give gifts to, creating a *GiftPair* for each combination.  Calculate the penalty 
associated that pairing.  The penalty would be 0 (zero) if that pairing has never been used in the past
(the giver has never gifted the recipient before).  If the pairing has been used before the penalty is based on how
many times and how long ago that pairing has been used.  There is also an additional penalty based on how often the
person has had to repeat gift (to anyone).

We now have an array consisting of every person, each having an array of possible recipient parings
(along with the penalty associated with that paring).  Note: not all persons have the same number of possible
recipients, so it's not a nice symmetrical 2 dimensional array.

You get a solution by selecting a possible paring from each person, adding up the penalties from each to get the
solutions's score.  Of course not all solutions are legal (a duplicate recipient, or a reciprocal).

###### A solution tree
From Wikipedia
> A tree is a (possibly non-linear) data structure made up of nodes or vertices and edges without having any cycle.
> A tree that is not empty consists of a root node and potentially many levels of additional nodes that form a hierarchy.

We start with the root node of our tree (an abstract notion), the children of this root node are all of the
possible gift parings of the first person in our person array.  Take each of these child nodes, their children
are all of the possible gift parings of the next person in our array.  (Each of these parings will be child nodes
of every one of the nodes from the previous level.)  Repeat for all persons in our array.

When completed you'll have:
*   A data structure representing all possible solutions
    *   Each node in the tree represents a partial solution (except for the lowest level nodes that have no
    children, that is a complete solution)
    * There is only one path from each node back to the root node.  So that part of the solution is well defined
    * Many possible solutions branch off from each node
    * Each complete path represents a solution
    *   Not all paths (solutions) are legal
*   A data structure that will take years to build and will not fit on your computer

**Why a tree?**  As we consider a partial solution represented by a particular node, we can ask a couple of questions;
Is that partial solution legal (See rules above)?  Is this partial solution better than the best (lowest score) complete
solution we've found so far?  If the answer to either of these questions is NO then we can eliminate this partial
solution AND and of the other solutions (partial and complete) that branch off of it.  **Pruning the tree!**

**What about the fact that we couldn't possibly build this wonderful tree?** We don't, the tree is a data
abstraction and we will look at it one path (node) at a time.



So how well does it work?  With a population of 22 folks.

```
Families                  :                                         4
Population                :                                        22
Possible Pairs            :                                       338
Possible Solutions        :       101,044,962,002,466,589,009,510,400
Elapsed Time              :                           32m 56.6356438s
Considered                :                           557,072,520,053
Already In Solution       :                           485,748,624,785
Reciprocal                :                             1,853,040,328
Score Too High            :                            26,752,813,773
Accepted                  :                            42,718,041,167
Solutions considered      :                                       111
```

```
Considered  : number of nodes (partial solutions) looked at
In Solution : number of nodes rejected because the recipient was already in the solution (illegal)
Reciprocal  : number of nodes rejected because the reciprocal pair was already in the solution (illegal)
Score       : number of nodes rejected because the score could not beat the best so far
Accepted    : number of nodes accepted for the next step (calling the recursive method)
Solutions   : number of complete solutions considered (111 out of 101,044,962,002,466,589,009,510,400)
```


33 minuets, not bad considering that a brute force method would take thousands of years!  But we can do better.
The ideal solution would be to have the first solution considered be the best and have every other solution pruned
as early as possible.  So if we want to look at the nodes (gift pairs) that have the highest penalties first.  So
sort the person list so the highest penalties are at the top.

We have:

```
Families                  :                                         4
Population                :                                        22
Possible Pairs            :                                       338
Possible Solutions        :       101,044,962,002,466,589,009,510,400
Elapsed Time              :                                 59.0059ms
Considered                :                                12,949,375
Already In Solution       :                                10,603,583
Reciprocal                :                                   143,535
Score Too High            :                                 1,253,772
Accepted                  :                                   948,485
Solutions considered      :                                        16
```

Less than one second!  This has been accomplished by pruning the tree much closer to the top.



