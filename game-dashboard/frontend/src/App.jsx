import { useState, useEffect } from 'react';
import { Table, Tag, Button, Alert } from 'antd'; // Alert 컴포넌트 추가

function App() {
  const [tickets, setTickets] = useState([]);
  const [error, setError] = useState(null); // 에러 상태 추가

  const fetchTickets = () => {
    setError(null); // 재시도 시 에러 초기화
    
    fetch('api/tickets')
      .then(res => {
        if (!res.ok) { // 404나 500 에러 체크
            throw new Error(`서버 에러 발생: ${res.status}`);
        }
        return res.json();
      })
      .then(data => {
        console.log("받은 데이터:", data); // F12 콘솔에서 데이터 확인용

        // 안전 장치: 데이터가 배열(리스트)인지 확인
        if (Array.isArray(data)) {
            setTickets(data);
        } else {
            // 데이터가 null이거나 이상하면 빈 배열로 처리
            setTickets([]);
            console.warn("데이터가 배열이 아닙니다:", data);
        }
      })
      .catch(err => {
        console.error("Fetch 에러:", err);
        setError(err.message); // 에러 메시지 저장
        setTickets([]); // 에러나면 빈 표 보여주기
      });
  };

  useEffect(() => {
    fetchTickets();
  }, []);

  const columns = [
    {
      title: 'Ticket ID',
      dataIndex: 'id',
      key: 'id',
    },
    {
      title: 'Search Fields',
      dataIndex: 'search_fields',
      key: 'search_fields',
      render: (fields) => {
        // fields가 없을 수도 있으니 안전하게 체크 (?.)
        return (
            <>
              {fields?.tags?.map(tag => <Tag color="blue" key={tag}>{tag}</Tag>)}
              {fields?.double_args?.mmr && <Tag color="green">MMR: {fields.double_args.mmr}</Tag>}
            </>
        );
      },
    },
    {
      title: 'Created At',
      dataIndex: 'create_time',
      key: 'create_time',
      render: (time) => {
          // time이 없을 경우 처리
          if (!time || !time.seconds) return "-";
          return new Date(time.seconds * 1000).toLocaleString();
      },
    },
  ];

  return (
    <div style={{ padding: 50 }}>
      <h1>🎾 Open Match Dashboard</h1>
      
      {/* 에러가 있으면 빨간 박스 보여주기 */}
      {error && (
        <Alert 
            message="데이터 불러오기 실패" 
            description={error} 
            type="error" 
            showIcon 
            style={{ marginBottom: 20 }}
        />
      )}

      <Button type="primary" onClick={fetchTickets} style={{ marginBottom: 16 }}>
        새로고침
      </Button>
      
      <Table 
        dataSource={tickets} 
        columns={columns} 
        rowKey="id" 
        locale={{ emptyText: '데이터가 없습니다 (서버 연결 확인 필요)' }} 
      />
    </div>
  );
}

export default App;
