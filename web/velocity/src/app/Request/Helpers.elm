module Request.Helpers exposing (apiUrl)

import HttpBuilder
import Http


apiUrl : String -> String
apiUrl str =
    "http://localhost/v1" ++ str
