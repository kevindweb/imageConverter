package main

/*
split image into M equally sized chunks

go routine to execute findIcon on each chunk
- start without multi-threading but convert once tests are valid

merge each chunk by scanning the edges for boundary icons
- if there's a hit, add these two components together in an array
    - merge the height and width
- find the biggest combined component

run build image on merged components
- check if pixel component is in the list of grouped icon chunks

examples:
https://github.com/bonej-org/BoneJ2/blob/da5aa63cdc15516605e8dcb77458eb34b0f00b85/Legacy/bonej/src/main/java/org/bonej/plugins/ConnectedComponents.java#L540
https://github.com/opencv/opencv/issues/7270

It's called multithreaded connected component labeling

*/
