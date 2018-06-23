package utils

// SwitchString is Ternary operator for strings
func SwitchString(cond bool, strIfTrue string, strIfFalse string) string {
	if cond {
		return strIfTrue
	}
	return strIfFalse

}
