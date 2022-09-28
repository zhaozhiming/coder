export const formatTemplateActiveDevelopers = (num?: number): string => {
  if (num === undefined || num < 0) {
    // Loading
    return "0"
  }
  return num.toString()
}
