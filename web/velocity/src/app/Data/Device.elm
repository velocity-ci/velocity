module Data.Device
    exposing
        ( Size
        , isLarge
        , isSmall
        , size
        )


type Size
    = Small
    | Large
    | ExtraLarge


size : Int -> Size
size pixelWidth =
    if pixelWidth > 1280 then
        ExtraLarge
    else if pixelWidth > 991 then
        Large
    else
        Small


isLarge : Size -> Bool
isLarge size =
    case size of
        ExtraLarge ->
            True

        Large ->
            True

        _ ->
            False


isSmall : Size -> Bool
isSmall size =
    not (isLarge size)
