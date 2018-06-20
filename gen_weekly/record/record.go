package record

type Record struct {
	Id          string `csv:"id" json:"id"`
	Name        string `csv:"name" json:"name"`
	Price       string `csv:"price" json:"price"`
	Category    string `csv:"category" json:"category"`
	DayStart    string `csv:"day_start" json":day_start"`
	DayEnd      string `csv:"day_end" json:"day_end"`
	CanWeekday  string `csv:"can_weekday" json:"can_weekday"`
	Description string `csv:"description" json:"description"`
}

func (r *Record) MarshalString() string {
	return r.Id + "," +
		r.Name + "," +
		r.Price + "," +
		r.Category + "," +
		r.DayStart + "," +
		r.DayEnd + "," +
		r.CanWeekday + "," +
		r.Description
}

func (r *Record) MarshalStringSlice() []string {
	return []string{
		r.Id,
		r.Name,
		r.Price,
		r.Category,
		r.DayStart,
		r.DayEnd,
		r.CanWeekday,
		r.Description}
}
