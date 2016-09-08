#!/usr/bin/env Rscript
LinearRegretion <- function(points){
	  require(jsonlite)
  frame <- fromJSON(points)
    
    fit2b <- lm(frame$y ~ frame$x, data = frame)
    
    alpha <- fit2b$coefficient[2]
      beta <- fit2b$coefficient[1]
      r2 <- summary(fit2b)$adj.r.squared
        
        retorno <- data.frame (alpha, beta, r2)
        toJSON(retorno)
}
args = commandArgs(trailingOnly=TRUE)
LinearRegretion(args[1])
#json <- '{"x":[3.87, 3.61, 4.33, 3.43, 3.81, 3.83, 3.46, 3.76,3.50, 3.58, 4.19, 3.78, 3.71, 3.73, 3.78],"y":[4.87, 3.93, 6.46, 3.33, 4.38, 4.70, 3.50, 4.50,3.58, 3.64, 5.90, 4.43, 4.38, 4.42, 4.25]}'
#LinearRegretion(json)

