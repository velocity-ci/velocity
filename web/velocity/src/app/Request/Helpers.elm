module Request.Helpers exposing (apiUrl)


apiUrl : { r | apiUrlBase : String } -> String -> String
apiUrl { apiUrlBase } str =
    apiUrlBase ++ str
