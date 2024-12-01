package engine

// Negligible float value for comparisons
// To check if a float is close to zero and can be considered zero
// For example to remove an impulse if it has decayed to negligible values
const negligibleFloat = 0.01
