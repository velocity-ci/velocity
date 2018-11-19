module Element.Link exposing (..)


link :
    List (Attribute msg)
    ->
        { url : String
        , label : Element msg
        }
    -> Element msg
