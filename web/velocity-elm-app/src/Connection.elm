module Connection exposing (Connection)

import Edge exposing (Edge)
import PageInfo exposing (PageInfo)


type alias Connection a =
    { pageInfo : PageInfo
    , edges : List (Edge a)
    }
