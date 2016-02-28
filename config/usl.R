#!/usr/bin/env Rscript
usl <- function(points){
  require(jsonlite)
  frame <- fromJSON(points)
  p <- frame$x
  X1 = frame$y[1]
  X <- frame$y
  df = data.frame(p, X)
  df$RelCap <- df$X / X1 #Variable
  df$Efficiency <- df$RelCap / p
  df$Inverse <- df$p / df$RelCap
  df$Linearity <- df$p - 1
  df$Deviation <- (df$p / df$RelCap) - 1
  
  fit2b <- lm(Deviation ~ I(Linearity^2) + Linearity - 1, data = df)
  fit2b$coefficient[1]#kappa
  fit2b$coefficient[2]#sigma
  alpha <- fit2b$coefficient[2] - fit2b$coefficient[1]
  beta <- fit2b$coefficient[1]
  r2 <- summary(fit2b)$adj.r.squared
  N <- floor(sqrt((1 - alpha) / beta))
  retorno <- data.frame (alpha, beta, r2, N)
  toJSON(retorno)
}
args = commandArgs(trailingOnly=TRUE)
usl(args[1])
