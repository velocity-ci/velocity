module Page.NotFound exposing (view)

import Asset
import Element exposing (..)


-- VIEW


view : { title : String, content : Element msg }
view =
    { title = "Page Not Found"
    , content =
        Element.row []
            [ text "Not Found"
            , image [] { src = Asset.src Asset.error, description = "" }
            ]
    }
