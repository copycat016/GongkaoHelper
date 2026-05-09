import { useEffect, useMemo, useState } from "react";
import { Badge, Calendar, Card, Col, List, Row, Segmented, Space, Tag } from "antd";
import PageHeader from "../components/PageHeader";
import StatCard from "../components/StatCard";
import { getCalendarEvents } from "../api/calendar";

function CalendarPage() {
  const [events, setEvents] = useState([]);
  const [selectedDate, setSelectedDate] = useState(new Date());
  const [filter, setFilter] = useState("全部");

  const loadEvents = async (date = new Date()) => {
    const month = `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}`;
    setEvents((await getCalendarEvents({ month })) || []);
  };

  useEffect(() => {
    loadEvents();
  }, []);

  const eventsByDate = useMemo(() => {
    return filteredEvents(events, filter).reduce((map, item) => {
      const key = new Date(item.date).toDateString();
      map[key] = [...(map[key] || []), item];
      return map;
    }, {});
  }, [events, filter]);

  const summary = useMemo(() => {
    const visible = filteredEvents(events, filter);
    return {
      total: visible.length,
      plans: visible.filter((item) => item.type === "学习计划").length,
      reviews: visible.filter((item) => item.type === "错题复习").length,
      logs: visible.filter((item) => item.type === "番茄钟学习记录").length,
    };
  }, [events, filter]);

  const cellRender = (current) => {
    const date = current.toDate ? current.toDate() : current;
    const dayEvents = eventsByDate[date.toDateString()] || [];
    return (
      <div className="calendar-events">
        {dayEvents.slice(0, 3).map((event) => (
          <Badge key={event.id} status={badgeStatus(event.type)} text={event.title} />
        ))}
      </div>
    );
  };

  const selectedEvents = eventsByDate[selectedDate.toDateString()] || [];
  const upcomingEvents = useMemo(() => {
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    return filteredEvents(events, filter)
      .filter((item) => new Date(item.date) >= today)
      .sort((a, b) => new Date(a.date) - new Date(b.date))
      .slice(0, 8);
  }, [events, filter]);

  return (
    <div className="page-grid">
      <PageHeader eyebrow="Calendar" title="日历" desc="计划、复习和学习记录。" />
      <Row gutter={[16, 16]}>
        <Col xs={24} md={6}><StatCard label="本月事项" value={String(summary.total)} hint="当前筛选" /></Col>
        <Col xs={24} md={6}><StatCard label="计划" value={String(summary.plans)} hint="待安排" /></Col>
        <Col xs={24} md={6}><StatCard label="复习" value={String(summary.reviews)} hint="错题提醒" /></Col>
        <Col xs={24} md={6}><StatCard label="记录" value={String(summary.logs)} hint="番茄钟" /></Col>
      </Row>
      <Card className="glass-card" bordered={false}>
        <Segmented
          value={filter}
          onChange={setFilter}
          options={["全部", "学习计划", "错题复习", "番茄钟学习记录"]}
        />
      </Card>
      <Row gutter={[16, 16]}>
        <Col xs={24} xl={16}>
          <Card className="glass-card calendar-card" bordered={false}>
            <Calendar
              cellRender={cellRender}
              onPanelChange={(value) => loadEvents(value.toDate ? value.toDate() : value)}
              onSelect={(value) => setSelectedDate(value.toDate ? value.toDate() : value)}
            />
          </Card>
        </Col>
        <Col xs={24} xl={8}>
          <Space direction="vertical" size="middle" className="calendar-side">
            <Card className="glass-card" title="当天事项" bordered={false}>
              <List
                dataSource={selectedEvents}
                locale={{ emptyText: "当天暂无事件" }}
                renderItem={(item) => <CalendarEventItem item={item} />}
              />
            </Card>
            <Card className="glass-card" title="本月待办" bordered={false}>
              <List
                dataSource={upcomingEvents}
                locale={{ emptyText: "暂无待办" }}
                renderItem={(item) => <CalendarEventItem item={item} />}
              />
            </Card>
          </Space>
        </Col>
      </Row>
    </div>
  );
}

function CalendarEventItem({ item }) {
  return (
    <List.Item>
      <List.Item.Meta
        title={<Space><Badge status={badgeStatus(item.type)} /><span>{item.title}</span></Space>}
        description={<Space wrap><Tag>{item.type}</Tag><span>{new Date(item.date).toLocaleDateString("zh-CN")}</span><span>{item.status}</span></Space>}
      />
    </List.Item>
  );
}

function filteredEvents(events, filter) {
  if (filter === "全部") return events;
  return events.filter((item) => item.type === filter);
}

function badgeStatus(type) {
  if (type === "错题复习") return "warning";
  if (type === "学习计划") return "processing";
  return "success";
}

export default CalendarPage;
